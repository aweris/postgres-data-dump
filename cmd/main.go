package main

import (
	syslog "log"
	"os"

	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/spf13/pflag"
)

// Version represents the software version of the
// nolint:gochecknoglobals
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var (
		// logging
		logLevel  string
		logFormat string
		verbose   bool

		// other
		showVersion bool
	)

	flag := pflag.NewFlagSet("defaults", pflag.ExitOnError)

	// no need to sort flags in help
	flag.SortFlags = false

	// logging flags
	flag.StringVar(&logLevel, "log-level", log.LevelError, "log filtering level. ('error', 'warn', 'info', 'debug')")
	flag.StringVar(&logFormat, "log-format", log.FormatFmt, "log format to use. ('fmt', 'json')")
	flag.BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// other flags
	flag.BoolVar(&showVersion, "version", false, "Prints version info")

	// bind environment variables
	bindEnv(flag.Lookup("log-level"), "PDD_LOG_LEVEL")
	bindEnv(flag.Lookup("log-format"), "PDD_LOG_FORMAT")

	// Parse Options
	if err := flag.Parse(os.Args); err != nil {
		syslog.Fatalf("%#v", err)
	}

	// print application version
	if showVersion {
		syslog.Printf("Version    : %s\n", version)
		syslog.Printf("Git Commit : %s\n", commit)
		syslog.Printf("Build Date : %s\n", date)
		os.Exit(0)
	}

	// Enable log level debug for verbose
	if verbose {
		logLevel = log.LevelDebug
	}

	logger, err := log.NewLogger(logLevel, logFormat, "pdd")
	if err != nil {
		syslog.Fatalf("%#v", err)
	}

	logger.Debug("version", version, "git commit", commit, "build date", date)
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			syslog.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
