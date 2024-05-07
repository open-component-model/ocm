package configutils

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
)

func Configure(path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(config.DefaultContext(), path, fss...)
	return err
}

func ConfigureContext(ctxp config.ContextProvider, path string, fss ...vfs.FileSystem) error {
	_, err := utils.Configure(ctxp, path, fss...)
	return err
}
