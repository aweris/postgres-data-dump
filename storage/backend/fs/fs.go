package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aweris/postgres-data-dump/internal/helpers"
	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/pkg/errors"
)

const (
	defaultFileMode = 0755
)

// Backend is an file system implementation of the Backend.
type Backend struct {
	root   string
	logger log.Logger
}

// New creates a Backend backend.
func New(logger log.Logger, c Config) (*Backend, error) {
	if strings.TrimRight(path.Clean(c.Root), "/") == "" {
		return nil, errors.Errorf("empty or root path given, <%s> as root", c.Root)
	}

	if _, err := os.Stat(c.Root); err != nil {
		return nil, errors.Wrapf(err, "make sure volume is mounted, <%s> as root", c.Root)
	}

	logger.Debug("msg", "fs backend", "config", fmt.Sprintf("%#v", c))

	return &Backend{logger: logger, root: c.Root}, nil
}

// Put uploads contents of the given reader.
func (b *Backend) Put(ctx context.Context, p string, r io.Reader) error {
	fp, err := filepath.Abs(filepath.Clean(filepath.Join(b.root, p)))
	if err != nil {
		return errors.Wrapf(err, "invalid file path: %s", p)
	}

	errCh := make(chan error)

	go func() {
		defer close(errCh)

		dir := filepath.Dir(fp)
		if err := os.MkdirAll(dir, os.FileMode(defaultFileMode)); err != nil {
			errCh <- errors.Wrap(err, "can't create directory")
		}

		w, err := os.Create(fp)
		if err != nil {
			errCh <- errors.Wrapf(err, "can't create file %s", fp)
			return
		}

		defer helpers.CloseWithErrLogf(b.logger, w, "response body, close defer")

		if _, err := io.Copy(w, r); err != nil {
			errCh <- errors.Wrapf(err, "can't write contents of reader to a file %s", fp)
		}

		if err := w.Close(); err != nil {
			errCh <- errors.Wrapf(err, "can't close object %s", fp)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
