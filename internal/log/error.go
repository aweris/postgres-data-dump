package log

import "errors"

var (
	ErrUnexpectedLogLevel  = errors.New("unexpected log level")
	ErrUnexpectedLogFormat = errors.New("unexpected log format")
)
