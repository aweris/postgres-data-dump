package database

import "time"

// default values.
const (
	DefaultAddr        = "localhost:5432"
	DefaultDatabase    = "postgres"
	DefaultUser        = "postgres"
	DefaultPassword    = "postgres"
	DefaultMaxRetries  = 0
	DefaultDialTimeout = 5 * time.Second
	DefaultReadTimeout = 30 * time.Second
)

// Config contains database connection options.
type Config struct {
	Addr        string
	Database    string
	User        string
	Password    string
	EnableSSL   bool
	MaxRetries  int
	DialTimeout time.Duration
	ReadTimeout time.Duration
}
