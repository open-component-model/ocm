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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/errors"
)

// pluginHandler stores artifact blobs as OCIArtifacts.
type pluginHandler struct {
	plugin plugin.Plugin
	name   string
}

func New(p plugin.Plugin, name string) (download.Handler, error) {
	dd := p.GetDownloaderDescriptor(name)
	if dd == nil {
		return nil, errors.ErrUnknown(ppi.KIND_DOWNLOADER, name, p.Name())
	}

	return &pluginHandler{
		plugin: p,
		name:   name,
	}, nil
}

func (b *pluginHandler) Download(_ common.Printer, racc cpi.ResourceAccess, path string, _ vfs.FileSystem) (bool, string, error) {
	m, err := racc.AccessMethod()
	if err != nil {
		return true, "", err
	}

	r := accessio.NewOndemandReader(m)
	defer errors.PropagateError(&err, r.Close)

	return b.plugin.Download(b.name, r, racc.Meta().Type, m.MimeType(), path)
}
