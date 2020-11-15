package dump

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/template"

	"github.com/aweris/postgres-data-dump/database"
	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/pkg/errors"
)

// Dump templates.
const (
	dumpHeader = `
--
-- PostgreSQL database dump
--	

BEGIN;
`

	dumpSettings = `
SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

SET search_path = public, pg_catalog;
`

	tableHeader = `
--
-- Data for Name: %s; Type: TABLE DATA
--

COPY %s (%s) FROM stdin;
`

	dumpFooter = `
COMMIT;

--
-- PostgreSQL database dump complete
--
`
)

// Dumper provides functionality to dump database with given manifest configuration.
type Dumper interface {
	// creates database dump
	Dump(w io.Writer) error
}

type dumper struct {
	logger   log.Logger
	db       database.DB
	manifest *manifest
	nav      *navigator
}

// NewDumper creates Dumper instance.
func NewDumper(logger log.Logger, db database.DB, cfg Config) (Dumper, error) {
	manifest, err := loadManifest(logger, cfg.ManifestFile)
	if err != nil {
		logger.Error("msg", "failed to create exporter", "error", err)

		return nil, errors.Wrap(err, "failed to create exporter")
	}

	nav := newNavigator(logger, db, manifest)

	logger.Debug("msg", "create exporter instance", "manifest", cfg.ManifestFile)

	return &dumper{
		logger:   logger,
		db:       db,
		manifest: manifest,
		nav:      nav,
	}, nil
}

func (d *dumper) Dump(w io.Writer) error {
	// Print dump header
	if _, err := fmt.Fprint(w, dumpHeader); err != nil {
		d.logger.Error("msg", "failed to write dump header", "error", err)

		return err
	}

	// Print dump settings
	if _, err := fmt.Fprint(w, dumpSettings); err != nil {
		d.logger.Error("msg", "failed to write dump settings", "error", err)

		return err
	}

	// Print tables
	for d.nav.hasNext() {
		t, err := d.nav.next()
		if err != nil {
			d.logger.Error("msg", "can't fetch next table", "error", err)

			return err
		}

		// Wrap column names with quote
		quoted := make([]string, 0)
		for _, v := range t.Columns {
			quoted = append(quoted, strconv.Quote(v))
		}

		// Join columns
		cols := strings.Join(quoted, ", ")

		// Print table copy statement with stdin option
		if _, err := fmt.Fprintf(w, tableHeader, t.TableName, t.TableName, cols); err != nil {
			d.logger.Error("msg", "failed to write table header", "error", err)

			return err
		}

		// Print table data
		source, err := copyFrom(d.manifest, t)
		if err != nil {
			return err
		}

		if err := d.db.CopyTo(w, source); err != nil {
			return err
		}

		// Print table footer
		if _, err := fmt.Fprintln(w, `\.`); err != nil {
			d.logger.Error("msg", "failed to write table footer", "error", err)

			return err
		}

		// Print post actions
		for _, action := range t.PostActions {
			if _, err := fmt.Fprintf(w, "\n%s;\n", action); err != nil {
				d.logger.Error("msg", "failed to write table action", "action", action, "error", err)

				return err
			}
		}
	}

	// Print dump footer
	if _, err := fmt.Fprint(w, dumpFooter); err != nil {
		d.logger.Error("msg", "failed to write dump footer", "error", err)

		return err
	}

	return nil
}

// copyFrom returns prepared table statement from table name or rendered query.
func copyFrom(m *manifest, t *table) (string, error) {
	if t.Query == "" {
		return t.TableName, nil
	}

	// Create new template from query
	tmpl, err := template.New("query").Parse(t.Query)
	if err != nil {
		return "", err
	}

	// Render template with vars
	var out bytes.Buffer
	if err := tmpl.Execute(&out, m.Vars); err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s)", out.String()), nil
}
