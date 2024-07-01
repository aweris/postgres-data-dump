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
      --log-format string        log format to use. ('fmt', 'json') (default "fmt")
      --log-level string         log filtering level. ('error', 'warn', 'info', 'debug') (default "error")
  -v, --verbose                  verbose output
      --addr string              TCP host:port or Unix socket depending on Network (default "localhost:5432")
      --database string          Database name (default "postgres")
      --user string              Database user (default "postgres")
      --pass string              Database password (default "postgres")
      --dial-timeout duration    Dial timeout for establishing new connections (default 5s)
      --read-timeout duration    Timeout for socket reads. If reached, commands will fail (default 30s)
      --max-retry int            Maximum number of retries before giving up.
      --manifest-file string     Path to manifest file (default ".pdd.yaml")
      --backend string           storage backend to use (filesystem) (default "filesystem")
      --filesystem-root string   local filesystem root directory (default "/tmp/pdd")
      --version                  Prints version info
```


### Environment Variables

Environment variables sets property value if it's not set from arguments. CLI arguments has higher priority.

| Name | Property |
|:---|:---|
| PDD_LOG_LEVEL | `--log-format` |
| PDD_LOG_FORMAT | `--log-level` |
| PDD_ADDR | `--addr` |
| PDD_DATABASE | `--database` |
| PDD_USER | `--user` |
| PDD_PASS | `--pass` |
| PDD_DIAL_TIMEOUT | `--dial-timeout` |
| PDD_READ_TIMEOUT | `--read-timeout` |
| PDD_MAX_RETRY | `--max-retry` |
| PDD_MANIFEST_FILE | `--manifest-file` |
| PDD_BACKEND | `--backend` |
| PDD_FILESYSTEM_ROOT | `--filesystem-root` |

### Manifest file

The main difference between `pg_dump_sample` and `pg_dump(1)` is that
`pg_dump_sample` requires a manifest file describing how to dump the database.
The manifest file is a YAML file describing what tables to dump and how to dump
them.

A quick example:

    ---
    vars:
      # Condition to dump only certain users
      matching_user_id: "(users.id BETWEEN 1000 AND 2000)"

    tables:
      # Dump everything from table "consts"
      - table: consts

      # Dump only matching users
      - table: users
        query: "SELECT * FROM users WHERE {{.matching_user_id}}"
        post_actions:
          - "SELECT pg_catalog.setval('users_id_seq', MAX(id) + 1, true) FROM users"

      # Dump only tickets that were bought by matching users
      - table: tickets
        query: >
          SELECT purchases.* FROM purchases, users
          WHERE
            purchases.buyer_id = users.id
            AND {{.matching_user_id}}


Currently, these top-level keys are available:

#### `vars`

Definitions of variables which will be used to replace placeholders in queries.

#### `tables`

List of tables to dump. Tables are dumped in the order they are specified in the
manifest file, with one exception: if the table contains foreign keys
referencing another table, the referenced table will be dumped first. This is to
ensure that the dump can be loaded later without errors.

By default, all rows of the table will be dumped. If you don't want to dump all
the rows use the `query` to specify a SELECT SQL statement which returns the
rows you want to dump.

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
