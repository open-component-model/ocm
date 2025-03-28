package configutils

import (
	_ "ocm.software/ocm/api/datacontext/config"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/config"
	utils "ocm.software/ocm/api/ocm/ocmutils"
)

// Configure configures the default context applying
// a configuration file.
// It also applies implicit settings from the OCM context.
func Configure(path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(config.DefaultContext(), path, fss...)
	return err
}

// ConfigureContext configures the given context applying
// a configuration file.
// It also applies implicit settings from the OCM context.
func ConfigureContext(ctxp config.ContextProvider, path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(ctxp, path, fss...)
	return err
}
