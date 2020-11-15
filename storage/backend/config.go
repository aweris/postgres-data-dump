package backend

import "github.com/aweris/postgres-data-dump/storage/backend/fs"

// Config configures behavior of Backend.
type Config struct {
	Type string

	FileSystem fs.Config
}
