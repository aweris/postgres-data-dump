package storage

import (
	"context"
	"io"
	"time"

	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/aweris/postgres-data-dump/storage/backend"
)

const (
	DefaultOperationTimeout = 3 * time.Minute
)

// Storage is a place that files can be written to and read from.
type Storage interface {
	// Put writes contents of io.Reader to remote storage at given key location.
	Put(p string, r io.Reader) error
}

// Default Storage implementation.
type storage struct {
	logger  log.Logger
	backend backend.Backend
	timeout time.Duration
}

// New create a new default storage.
func New(logger log.Logger, b backend.Backend, timeout time.Duration) Storage {
	return &storage{logger, b, timeout}
}

// Put writes contents of io.Reader to remote storage at given key location.
func (s *storage) Put(p string, r io.Reader) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	return s.backend.Put(ctx, p, r)
}
