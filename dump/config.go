package dump

// default values.
const (
	DefaultManifestFile = ".pdd.yaml"
)

// Config contains export configuration options.
type Config struct {
	ManifestFile string
}
