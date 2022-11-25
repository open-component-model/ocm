// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type Config = json.RawMessage

func init() {
	cpi.RegisterBlobHandlerRegistrationHandler("plugin", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ cpi.BlobHandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config cpi.BlobHandlerConfig, olist ...cpi.BlobHandlerOption) (bool, error) {
	path := cpi.NewNamePath(handler)

	if config == nil {
		return true, fmt.Errorf("target specification required")
	}

	if len(path) < 1 || len(path) > 2 {
		return true, fmt.Errorf("plugin handler must be of the form <plugin>[/<uploader>]")
	}

	opts := cpi.NewBlobHandlerOptions(olist...)

	var attr Config
	switch a := config.(type) {
	case json.RawMessage:
		attr = a
	case []byte:
		err := runtime.DefaultYAMLEncoding.Unmarshal(a, &attr)
		if err != nil {
			return true, errors.Wrapf(err, "invalid target specification")
		}
		attr = a
	default:
		data, err := json.Marshal(config)
		if err != nil {
			return true, errors.Wrapf(err, "invalid target specification")
		}
		attr = data
	}

	name := ""
	if len(path) > 1 {
		name = path[1]
	}
	_, _, err := RegisterBlobHandler(ctx, path[0], name, opts.ArtifactType, opts.MimeType, attr)
	return true, err
}

func RegisterBlobHandler(ctx ocm.Context, pname, name string, artType, mediaType string, target json.RawMessage) (string, plugin.UploaderKeySet, error) {
	set := plugincacheattr.Get(ctx)
	if set == nil {
		return "", nil, errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	p := set.Get(pname)
	if p == nil {
		return "", nil, errors.ErrUnknown(plugin.KIND_PLUGIN, pname)
	}

	if name != "" {
		if p.GetUploaderDescriptor(name) == nil {
			return "", nil, fmt.Errorf("uploader %s not found in plugin %q", name, pname)
		}
	}
	keys := plugin.UploaderKeySet{}.Add(plugin.UploaderKey{}.SetArtifact(artType, mediaType))
	d := p.LookupUploader(name, artType, mediaType)

	if len(d) == 0 {
		keys = p.LookupUploaderKeys(name, artType, mediaType)
		if len(keys) == 0 {
			if name == "" {
				return "", nil, fmt.Errorf("no uploader found for [art:%q, media:%q]", artType, mediaType)
			}
			return "", nil, fmt.Errorf("uploader %s not valid for [art:%q, media:%q]", name, artType, mediaType)
		}
		d = p.LookupUploadersForKeys(name, keys)
	}
	if len(d) > 1 {
		return "", nil, fmt.Errorf("multiple uploaders found for [art:%q, media:%q]: %s", artType, mediaType, strings.Join(d.GetNames(), ", "))
	}
	h, err := New(p, d[0].Name, target)
	if err != nil {
		return d[0].Name, nil, err
	}
	for k := range keys {
		ctx.BlobHandlers().Register(h, cpi.ForArtifactType(k.GetArtifactType()), cpi.ForMimeType(k.GetMediaType()))
	}
	return d[0].Name, keys, nil
}
