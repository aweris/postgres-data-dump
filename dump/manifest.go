package dump

import (
	"io/ioutil"
	"os"

	"github.com/aweris/postgres-data-dump/internal/log"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// manifest contains configuration describing how to export the database.
type manifest struct {
	Vars   map[string]string `yaml:"vars"`
	Tables []table           `yaml:"tables"`
}

// table contains table configuration for the export.
type table struct {
	TableName   string   `yaml:"table"`
	Query       string   `yaml:"query"`
	Columns     []string `yaml:"columns,flow"`
	PostActions []string `yaml:"post_actions,flow"`
}

// loadManifest creates new manifest instance from given file.
func loadManifest(logger log.Logger, manifestFile string) (*manifest, error) {
	// Open manifest file
	file, err := os.Open(manifestFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open manifest file")
	}

	// Read manifest file
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read manifest file")
	}

	// Unmarshal manifest
	manifest := manifest{}

	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal manifest file")
	}

	logger.Debug("msg", "load manifest file", "file", manifestFile)

	return &manifest, nil
}
