package helpers

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aweris/postgres-data-dump/internal/log"
)

// CloseWithErrLogf is making sure we log every error, even those from best effort tiny closers.
func CloseWithErrLogf(logger log.Logger, closer io.Closer, format string, a ...interface{}) {
	err := close(closer)
	if err == nil {
		return
	}

	logger.Error("msg", "detected close error", "err", fmt.Errorf(format+", %w", append(a, err)...))
}

//goland:noinspection GoReservedWordUsedAsName
func close(closer io.Closer) error {
	err := closer.Close()
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrClosed) {
		return nil
	}

	return err
}
