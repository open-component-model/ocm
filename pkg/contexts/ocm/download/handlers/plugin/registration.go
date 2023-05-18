// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/registrations"
)

type Config = json.RawMessage

func init() {
	download.RegisterHandlerRegistrationHandler("plugin", &RegistrationHandler{})
}

type RegistrationHandler struct{}

var _ download.HandlerRegistrationHandler = (*RegistrationHandler)(nil)

func (r *RegistrationHandler) RegisterByName(handler string, ctx cpi.Context, config download.HandlerConfig, olist ...download.HandlerOption) (bool, error) {
	path := cpi.NewNamePath(handler)

	if config == nil {
		return true, fmt.Errorf("target specification required")
	}

	if len(path) < 1 || len(path) > 2 {
		return true, fmt.Errorf("plugin handler name must be of the form <plugin>[/<downloader>]")
	}

	opts := download.NewHandlerOptions(olist...)

	name := ""
	if len(path) > 1 {
		name = path[1]
	}

	attr, err := registrations.DecodeAnyConfig(config)
	if err != nil {
		return true, errors.Wrapf(err, "plugin download handler config for %s/%s", path[0], name)
	}

	err = RegisterDownloadHandler(ctx, path[0], name, opts.ArtifactType, opts.MimeType, attr)
	return true, err
}

func RegisterDownloadHandler(ctx cpi.Context, pname, name string, artType, mediaType string, config []byte) error {
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
		if len(config) != 0 {
			if e.ConfigScheme == "" {
				return errors.Newf("no config accepted by download handler")
			}
			err := ValidateConfig([]byte(e.ConfigScheme), config)
			if err != nil {
				return err
			}
		}
		h, err := New(p, e.Name, config)
		if err != nil {
			return err
		}
		download.For(ctx).Register(artType, mediaType, h)
	}
	return nil
}

func ValidateConfig(schemadata, configdata []byte) error {
	if string(schemadata) == "any" {
		var i interface{}
		return json.Unmarshal(configdata, &i)
	}
	data, err := yaml.YAMLToJSON(schemadata)
	if err != nil {
		return errors.Wrapf(err, "invalid JSON scheme for downloader config")
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(configdata))
	if err != nil {
		return errors.Wrapf(err, "invalid JSON scheme for downloader config")
	}

	loader := gojsonschema.NewBytesLoader(data)
	res, err := schema.Validate(loader)
	if err != nil {
		return err
	}

	if !res.Valid() {
		errs := res.Errors()
		errMsg := errs[0].String()
		for i := 1; i < len(errs); i++ {
			errMsg = fmt.Sprintf("%s;%s", errMsg, errs[i].String())
		}
		return errors.New(errMsg)
	}
	return nil
}

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	infos := registrations.HandlerInfos{}

	set := plugincacheattr.Get(ctx)
	if set == nil {
		return infos
	}

	for _, name := range set.PluginNames() {
		for _, d := range set.Get(name).GetDescriptor().Downloaders {
			i := registrations.HandlerInfo{
				Name:        name + "/" + d.GetName(),
				ShortDesc:   "",
				Description: d.GetDescription(),
			}
			infos = append(infos, i)
		}
	}
	return infos
}
