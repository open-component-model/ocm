package plugin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/utils/registrations"
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

	name := ""
	if len(path) > 1 {
		name = path[1]
	}

	attr, err := registrations.DecodeAnyConfig(config)
	if err != nil {
		return true, errors.Wrapf(err, "plugin upload handler target config for %s/%s", path[0], name)
	}

	_, _, err = RegisterBlobHandler(ctx, path[0], name, opts.ArtifactType, opts.MimeType, attr)
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

func (r *RegistrationHandler) GetHandlers(ctx cpi.Context) registrations.HandlerInfos {
	infos := registrations.NewNodeHandlerInfo("downloaders provided by plugins",
		"sub namespace of the form <code>&lt;plugin name>/&lt;handler></code>")

	set := plugincacheattr.Get(ctx)
	if set == nil {
		return infos
	}

	for _, name := range set.PluginNames() {
		p := set.Get(name)
		if !p.IsValid() {
			continue
		}
		for _, u := range p.GetDescriptor().Uploaders {
			i := registrations.HandlerInfo{
				Name:        name + "/" + u.GetName(),
				ShortDesc:   "",
				Description: u.GetDescription(),
			}
			infos = append(infos, i)
		}
	}
	return infos
}
