package dockerconfig

import (
	dockercli "github.com/docker/cli/cli/config"
	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/defaultconfigregistry"
)

func init() {
	defaultconfigregistry.RegisterDefaultConfigHandler(DefaultConfigHandler, desc)
}

func DefaultConfigHandler(cfg config.Context) (string, config.Config, error) {
	// use docker config as default config for ocm cli
	d := filepath.Join(dockercli.Dir(), dockercli.ConfigFileName)
	if ok, err := vfs.FileExists(osfs.New(), d); ok && err == nil {
		ccfg := credcfg.New()
		ccfg.AddRepository(NewRepositorySpec(d, true))
		return d, ccfg, nil
	}
	return "", nil, nil
}

var desc = `
The docker configuration file at <code>~/.docker/config.json</code> is
read to feed in the configured credentials for OCI registries.
`
