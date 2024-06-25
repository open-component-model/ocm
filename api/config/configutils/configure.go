package configutils

import (
	_ "github.com/open-component-model/ocm/api/datacontext/config"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/api/config"
	utils "github.com/open-component-model/ocm/api/ocm/ocmutils"
)

func Configure(path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(config.DefaultContext(), path, fss...)
	return err
}

func ConfigureContext(ctxp config.ContextProvider, path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(ctxp, path, fss...)
	return err
}
