package npm

import (
	"fmt"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/config"
	credcfg "ocm.software/ocm/api/credentials/config"
	"ocm.software/ocm/api/ocm/ocmutils/defaultconfigregistry"
)

const (
	ConfigFileName = ".npmrc"
)

func init() {
	defaultconfigregistry.RegisterDefaultConfigHandler(DefaultConfigHandler, desc)
}

func DefaultConfig() (string, error) {
	d, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, ConfigFileName), nil
}

func DefaultConfigHandler(cfg config.Context) (string, config.Config, error) {
	// use docker config as default config for ocm cli
	d, err := DefaultConfig()
	if err != nil {
		return "", nil, nil
	}
	if ok, err := vfs.FileExists(osfs.OsFs, d); ok && err == nil {
		ccfg := credcfg.New()
		ccfg.AddRepository(NewRepositorySpec(d, true))
		return d, ccfg, nil
	}
	return "", nil, nil
}

var desc = fmt.Sprintf(`
The npm configuration file at <code>~/%s</code> is
read to feed in the configured credentials for NPM registries.
`, ConfigFileName)
