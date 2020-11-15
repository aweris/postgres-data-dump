package fs

const (
	DefaultRoot = "/tmp/pdd"
)

// Config is a structure to store filesystem backend configuration.
type Config struct {
	Root string
}
