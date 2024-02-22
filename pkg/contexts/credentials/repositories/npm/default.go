package npm

import (
	"fmt"
	"os"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/defaultconfigregistry"
	"github.com/open-component-model/ocm/pkg/errors"
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

func DefaultConfigHandler(cfg config.Context) error {
	// use docker config as default config for ocm cli
	d, err := DefaultConfig()
	if err != nil {
		return nil
	}
	if ok, err := vfs.FileExists(osfs.New(), d); ok && err == nil {
		ccfg := credcfg.New()
		ccfg.AddRepository(NewRepositorySpec(d, true))
		err = cfg.ApplyConfig(ccfg, d)
		if err != nil {
			return errors.Wrapf(err, "cannot apply npm config %q", d)
		}
	}
	return nil
}

var desc = fmt.Sprintf(`
The npm configuration file at <code>~/%s</code> is
read to feed in the configured credentials for NPM registries.
`, ConfigFileName)
