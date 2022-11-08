// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"encoding/json"
	"fmt"

	blobhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/generic/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	downhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/errors"
)

func (pi *pluginsImpl) RegisterBlobHandler(pname, name string, artType, mediaType string, target json.RawMessage) error {
	p := pi.Get(pname)
	if p == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	d := p.LookupUploader(name, artType, mediaType)
	if len(d) == 0 {
		if name == "" {
			return fmt.Errorf("no uploader found for [art:%q, media:%q]", artType, mediaType)
		}
		return fmt.Errorf("uploader %s not valid for [art:%q, media:%q]", name, artType, mediaType)
	}
	for _, e := range d {
		h, err := blobhdlr.New(p, e.Name, target)
		if err != nil {
			return err
		}
		pi.ctx.BlobHandlers().Register(h, cpi.ForArtefactType(artType), cpi.ForMimeType(mediaType))
	}
	return nil
}

func (pi *pluginsImpl) RegisterDownloadHandler(pname, name string, artType, mediaType string) error {
	p := pi.Get(pname)
	if p == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	d := p.LookupDownloader(name, artType, mediaType)
	if len(d) == 0 {
		if name == "" {
			return fmt.Errorf("no downloader found for [art:%q, media:%q]", artType, mediaType)
		}
		return fmt.Errorf("downloader %s not valid for [art:%q, media:%q]", name, artType, mediaType)
	}
	for _, e := range d {
		h, err := downhdlr.New(p, e.Name)
		if err != nil {
			return err
		}
		download.For(pi.ctx).Register(artType, mediaType, h)
	}
	return nil
}
