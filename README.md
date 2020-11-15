# Postgres Data Dump

This is a simple tool for dumping a sample of data from a PostgreSQL database.
The resulting dump can be loaded back into the new database using standard tools
(e.g. `psql(1)`). The original project belongs to: [dankeder/pg_dump_sample](https://github.com/dankeder/pg_dump_sample)

## Why would I want this?

It is useful if you have a huge PostgreSQL database, and you need a
database with a small dataset for testing or development.

## Usage

### CLI

```
Usage of pdd:
      --log-format string   log format to use. ('fmt', 'json') (default "fmt")
      --log-level string    log filtering level. ('error', 'warn', 'info', 'debug') (default "error")
  -v, --verbose             verbose output
      --version             Prints version info
```

## Development 

```

Usage:
  make <target>

Targets:
  help                    shows this help message
  fix                     Fix found issues (if it's supported by the $(GOLANGCI_LINT))
  fmt                     Runs gofmt
  lint                    Runs golangci-lint analysis
  clean                   Cleanup everything
  vendor                  Updates vendored copy of dependencies
  build                   Builds binary
  version                 Shows application version
  test                    Runs go test
```
