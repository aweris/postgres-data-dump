package main

import (
	"fmt"
	"io"
	syslog "log"
	"os"
	"time"

	"github.com/aweris/postgres-data-dump/database"
	"github.com/aweris/postgres-data-dump/dump"
	"github.com/aweris/postgres-data-dump/internal/helpers"
	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/aweris/postgres-data-dump/storage"
	"github.com/aweris/postgres-data-dump/storage/backend"
	"github.com/aweris/postgres-data-dump/storage/backend/fs"
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

		// database
		dbc = database.Config{}

		// dump
		dc = dump.Config{}

		// backend
		bc = backend.Config{}

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

	// database flags
	flag.StringVar(&dbc.Addr, "addr", database.DefaultAddr, "TCP host:port or Unix socket depending on Network")
	flag.StringVar(&dbc.Database, "database", database.DefaultDatabase, "Database name")
	flag.StringVar(&dbc.User, "user", database.DefaultUser, "Database user")
	flag.StringVar(&dbc.Password, "pass", database.DefaultPassword, "Database password")
	flag.DurationVar(&dbc.DialTimeout, "dial-timeout", database.DefaultDialTimeout, "Dial timeout for establishing new connections")
	flag.DurationVar(&dbc.ReadTimeout, "read-timeout", database.DefaultReadTimeout, "Timeout for socket reads. If reached, commands will fail")
	flag.IntVar(&dbc.MaxRetries, "max-retry", database.DefaultMaxRetries, "Maximum number of retries before giving up.")

	// dump flags
	flag.StringVar(&dc.ManifestFile, "manifest-file", dump.DefaultManifestFile, "Path to manifest file")

	// backend
	flag.StringVar(&bc.Type, "backend", backend.FileSystem, "storage backend to use (filesystem)")

	// backend filesystem
	flag.StringVar(&bc.FileSystem.Root, "filesystem-root", fs.DefaultRoot, "local filesystem root directory")

	// other flags
	flag.BoolVar(&showVersion, "version", false, "Prints version info")

	// bind environment variables
	bindEnv(flag.Lookup("log-level"), "PDD_LOG_LEVEL")
	bindEnv(flag.Lookup("log-format"), "PDD_LOG_FORMAT")

	// database variables
	bindEnv(flag.Lookup("addr"), "PDD_ADDR")
	bindEnv(flag.Lookup("database"), "PDD_DATABASE")
	bindEnv(flag.Lookup("user"), "PDD_USER")
	bindEnv(flag.Lookup("pass"), "PDD_PASS")
	bindEnv(flag.Lookup("dial-timeout"), "PDD_DIAL_TIMEOUT")
	bindEnv(flag.Lookup("read-timeout"), "PDD_READ_TIMEOUT")
	bindEnv(flag.Lookup("max-retry"), "PDD_MAX_RETRY")

	// dump variables
	bindEnv(flag.Lookup("manifest-file"), "PDD_MANIFEST_FILE")

	// backend variables
	bindEnv(flag.Lookup("backend"), "PDD_BACKEND")

	// backend - fs
	bindEnv(flag.Lookup("filesystem-root"), "PDD_FILESYSTEM_ROOT")

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

	// initialize db
	db, err := database.ConnectDB(logger, &dbc)
	if err != nil {
		logger.Error("msg", "failed to create database", "error", err)
		os.Exit(1)
	}

	// initialize dumper
	dumper, err := dump.NewDumper(logger, db, dc)
	if err != nil {
		logger.Error("msg", "failed to create dumper", "error", err)
		os.Exit(1)
	}

	// initialize backend
	b, err := backend.FromConfig(logger, bc)
	if err != nil {
		logger.Error("msg", "failed to create backend", "error", err)
		os.Exit(1)
	}

	// initialize storage
	s := storage.New(logger, b, storage.DefaultOperationTimeout)

	if err := run(logger, dumper, s); err != nil {
		logger.Error("msg", "failed to dump database", "error", err)
		os.Exit(1)
	}

	logger.Debug("msg", "export finished")
}

func run(logger log.Logger, dumper dump.Dumper, s storage.Storage) error {
	// create a synchronous in-memory pipe.
	pr, pw := io.Pipe()

	defer helpers.CloseWithErrLogf(logger, pr, "dump error")

	go func() {
		defer helpers.CloseWithErrLogf(logger, pw, "dump error")

		if err := dumper.Dump(pw); err != nil {
			logger.Error("error", err)
		}
	}()

	return s.Put(generateFileName(), pr)
}

// generateFileName generates new file name for dump based on timestamp.
func generateFileName() string {
	t := time.Now()
	return fmt.Sprintf("dump-%s.sql", t.Format("20060102-150405"))
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
