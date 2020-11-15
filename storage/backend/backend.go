package backend

import (
	"context"
	"io"

	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/aweris/postgres-data-dump/storage/backend/fs"
	"github.com/pkg/errors"
)

const (
	// FileSystem type of the corresponding backend represented as string constant.
	FileSystem = "filesystem"
)

// Backend implements operations for storage files.
type Backend interface {
	// Put uploads contents of the given reader.
	Put(ctx context.Context, p string, r io.Reader) error
}

// FromConfig creates new Backend by initializing  using given configuration.
func FromConfig(logger log.Logger, cfg Config) (Backend, error) {
	var (
		b   Backend
		err error
	)

	switch cfg.Type {
	case FileSystem:
		logger.Debug("msg", "using filesystem as backend")
		b, err = fs.New(logger.With("backend", FileSystem), cfg.FileSystem)
	default:
		return nil, ErrUnknownBackendType
	}

	if err != nil {
		return nil, errors.Wrapf(err, "can't initialize backend %s", cfg.Type)
	}

	return b, nil
}
