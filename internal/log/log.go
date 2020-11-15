package log

import (
	syslog "log"
	"os"

	kitLog "github.com/go-kit/kit/log"
	kitLevel "github.com/go-kit/kit/log/level"
)

const (
	// log formats.
	FormatFmt  = "fmt"
	FormatJSON = "json"

	// log levels.
	LevelError = "error"
	LevelWarn  = "warn"
	LevelInfo  = "info"
	LevelDebug = "debug"

	// depth of the caller file and line.
	// Since we wrapped logger instance we need to increase log.DefaultCaller's depth by one.
	loggerDepth = 4
)

// Logger is a wrapper interface for kitLeveled logging entries.
type Logger interface {
	With(keyvals ...interface{}) Logger

	Debug(keyvals ...interface{})
	Info(keyvals ...interface{})
	Warn(keyvals ...interface{})
	Error(keyvals ...interface{})
}

// logger simple wrapper object for log.Logger.
type logger struct {
	name     string
	logger   kitLog.Logger
	fallback func(err error, name string, kitLevel string, keyvals ...interface{})
}

func NewLogger(logLevel, logFormat, name string) (Logger, error) {
	var (
		kitLogger kitLog.Logger
		lvl       kitLevel.Option
	)

	switch logLevel {
	case LevelError:
		lvl = kitLevel.AllowError()
	case LevelWarn:
		lvl = kitLevel.AllowWarn()
	case LevelInfo:
		lvl = kitLevel.AllowInfo()
	case LevelDebug:
		lvl = kitLevel.AllowDebug()
	default:
		return nil, ErrUnexpectedLogLevel
	}

	switch logFormat {
	case FormatJSON:
		kitLogger = kitLog.NewJSONLogger(kitLog.NewSyncWriter(os.Stderr))
	case FormatFmt:
		kitLogger = kitLog.NewLogfmtLogger(kitLog.NewSyncWriter(os.Stderr))
	default:
		return nil, ErrUnexpectedLogFormat
	}

	kitLogger = kitLevel.NewFilter(kitLogger, lvl)
	kitLogger = kitLog.With(kitLogger, "name", name)
	kitLogger = kitLog.With(kitLogger, "ts", kitLog.DefaultTimestampUTC, "caller", kitLog.Caller(loggerDepth))

	return &logger{name: name, logger: kitLogger, fallback: fallbackLogger}, nil
}

// With returns a new contextual logger with keyvals prepended to those passed to calls to Log.
func (l *logger) With(keyvals ...interface{}) Logger {
	return &logger{name: l.name, logger: kitLog.With(l.logger, keyvals), fallback: l.fallback}
}

// Debug is a wrapper for kitLevel.Debug(logs).Log(keyvals).
func (l *logger) Debug(keyvals ...interface{}) {
	if err := kitLevel.Debug(l.logger).Log(keyvals...); err != nil {
		l.fallback(err, l.name, LevelDebug, keyvals)
	}
}

// Info is a wrapper for kitLevel.Info(logs).Log(keyvals).
func (l *logger) Info(keyvals ...interface{}) {
	if err := kitLevel.Info(l.logger).Log(keyvals...); err != nil {
		l.fallback(err, l.name, LevelInfo, keyvals)
	}
}

// Warn is a wrapper for kitLevel.Warn(logs).Log(keyvals).
func (l *logger) Warn(keyvals ...interface{}) {
	if err := kitLevel.Warn(l.logger).Log(keyvals...); err != nil {
		l.fallback(err, l.name, LevelWarn, keyvals)
	}
}

// Error is a wrapper for kitLevel.Error(logs).Log(keyvals).
func (l *logger) Error(keyvals ...interface{}) {
	if err := kitLevel.Error(l.logger).Log(keyvals...); err != nil {
		l.fallback(err, l.name, LevelError, keyvals)
	}
}

// fallbackLogger is a fallback syslog logger, in case of actual logger failures.
func fallbackLogger(err error, name string, kitLevel string, keyvals ...interface{}) {
	syslog.Printf("[fallback: %s] log: %v, kitLevel: %s, err: %v", name, keyvals, kitLevel, err)
}
