// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/errors"
)

func RegisterDownloadHandler(ctx ocm.Context, pname, name string, artType, mediaType string) error {
	set := plugincacheattr.Get(ctx)
	if set == nil {
		return errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	p := set.Get(pname)
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
		h, err := New(p, e.Name)
		if err != nil {
			return err
		}
		download.For(ctx).Register(artType, mediaType, h)
	}
	return nil
}
