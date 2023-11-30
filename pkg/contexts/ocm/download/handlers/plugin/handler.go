// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/finalizer"
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
	finalize.Close(r, "reader for downlowd download")

	return b.plugin.Download(b.name, r, racc.Meta().Type, m.MimeType(), path, b.config)
}
