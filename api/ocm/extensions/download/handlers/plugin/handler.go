package plugin

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/utils/accessio"
	common "ocm.software/ocm/api/utils/misc"
)

// pluginHandler delegates download format of artifacts to a plugin based handler.
type pluginHandler struct {
	plugin plugin.Plugin
	name   string
	config []byte
}

func New(p plugin.Plugin, name string, config []byte) (download.Handler, error) {
	dd := p.GetDownloaderDescriptor(name)
	if dd == nil {
		return nil, errors.ErrUnknown(descriptor.KIND_DOWNLOADER, name, p.Name())
	}

	return &pluginHandler{
		plugin: p,
		name:   name,
		config: config,
	}, nil
}

func (b *pluginHandler) Download(_ common.Printer, racc cpi.ResourceAccess, path string, _ vfs.FileSystem) (resp bool, eff string, rerr error) {
	m, err := racc.AccessMethod()
	if err != nil {
		return true, "", err
	}
	var finalize finalizer.Finalizer
	defer finalize.FinalizeWithErrorPropagation(&rerr)

	finalize.Close(m, "method for download")
	r := accessio.NewOndemandReader(m)
	finalize.Close(r, "reader for download")

	return b.plugin.Download(b.name, r, racc.Meta().Type, m.MimeType(), path, b.config)
}
