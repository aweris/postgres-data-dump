package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"strings"

	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/go-pg/pg/extra/pgotel"
	"github.com/go-pg/pg/v10"
	"github.com/pkg/errors"
)

// DB wrapper interface for the postgres database.
type DB interface {
	// GetTableColumns returns column names for the given table
	GetTableColumns(table string) ([]string, error)

	// GetTableDependencies returns  dependent tables for the given table
	GetTableDependencies(table string) ([]string, error)

	// CopyTo copy data from a table to io.Writer
	CopyTo(w io.Writer, table string) error
}

type db struct {
	pgdb   *pg.DB
	logger log.Logger
}

// ConnectDB connects to a database using provided options.
func ConnectDB(logger log.Logger, cfg *Config) (DB, error) {
	var tlsConfig tls.Config

	if cfg.EnableSSL {
		serverName := strings.Split(cfg.Addr, ":")[0]
		tlsConfig = tls.Config{ServerName: serverName}
	}

	pgdb := pg.Connect(
		&pg.Options{
			Addr:        cfg.Addr,
			User:        cfg.User,
			Password:    cfg.Password,
			Database:    cfg.Database,
			MaxRetries:  cfg.MaxRetries,
			DialTimeout: cfg.DialTimeout,
			ReadTimeout: cfg.ReadTimeout,
			TLSConfig:   &tlsConfig,
		},
	)

	// Enable tracing
	pgdb.AddQueryHook(pgotel.TracingHook{})

	if err := pgdb.Ping(context.Background()); err != nil {
		return nil, errors.Wrap(err, "failed to connect database")
	}

	logger.Debug("msg", "connected to the database", "database", cfg.Database, "user", cfg.User)

	return &db{pgdb: pgdb, logger: logger}, nil
}

func (d *db) GetTableColumns(table string) ([]string, error) {
	var model []struct{ Name string }

	sql := `
		SELECT attname as name
		FROM pg_catalog.pg_attribute
		WHERE attrelid = ?::regclass
		  AND attnum > 0
		  AND attisdropped = FALSE
		ORDER BY attnum
	`

	if _, err := d.pgdb.Query(&model, sql, table); err != nil {
		d.logger.Error("msg", "failed to get table columns", "table", table, "err", err)

		return nil, errors.Wrap(err, "failed to get table columns")
	}

	cols := make([]string, 0)
	for _, v := range model {
		cols = append(cols, v.Name)
	}

	d.logger.Debug("msg", "get table columns", "table", table, "cols", strings.Join(cols, ","))

	return cols, nil
}

func (d *db) GetTableDependencies(table string) ([]string, error) {
	var model []struct{ Name string }

	sql := `
		SELECT confrelid::regclass AS name
		FROM pg_catalog.pg_constraint
		WHERE conrelid = ?::regclass
		  AND contype = 'f'
	`

	if _, err := d.pgdb.Query(&model, sql, table); err != nil {
		d.logger.Error("msg", "failed to get table dependencies", "table", table, "err", err)

		return nil, errors.Wrap(err, "failed to get table dependencies")
	}

	tables := make([]string, 0)
	for _, v := range model {
		tables = append(tables, v.Name)
	}

	d.logger.Debug("msg", "get table dependencies", "table", table, "dependencies", strings.Join(tables, ","))

	return tables, nil
}

func (d *db) CopyTo(w io.Writer, table string) error {
	if _, err := d.pgdb.CopyTo(w, fmt.Sprintf("COPY %s TO STDOUT", table)); err != nil {
		return err
	}

	return nil
}
