// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugins

import (
	"encoding/json"
	"fmt"

	blobhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/generic/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/errors"
)

func (pi *pluginsImpl) RegisterBlobHandler(pname, name string, artType, mediaType string, target json.RawMessage) error {
	p := pi.Get(pname)
	if p == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}
	d := p.LookupUploader(name, artType, mediaType)
	if d == nil {
		if name == "" {
			return fmt.Errorf("no uploader found for [art:%q, media:%q]", artType, mediaType)
		}
		return fmt.Errorf("uploader %s not valid for [art:%q, media:%q]", name, artType, mediaType)
	}
	h, err := blobhdlr.New(p, d.Name, target)
	if err != nil {
		return err
	}
	pi.ctx.BlobHandlers().Register(h, cpi.ForArtefactType(artType), cpi.ForMimeType(mediaType))
	return nil
}
