// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"encoding/json"
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/registry"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/runtime"
)

type plugin struct {
	name       string
	version    string
	descriptor descriptor.Descriptor
	tweaker    func(descriptor descriptor.Descriptor) descriptor.Descriptor
	options    Options

	downloaders  map[string]Downloader
	downmappings *registry.Registry[Downloader, DownloaderKey]

	uploaders      map[string]Uploader
	upmappings     *registry.Registry[Uploader, UploaderKey]
	uploaderScheme runtime.Scheme

	methods      map[string]AccessMethod
	accessScheme runtime.Scheme

	actions map[string]Action

	configParser func(message json.RawMessage) (interface{}, error)
}

func NewPlugin(name string, version string) Plugin {
	var rt runtime.VersionedTypedObject
	return &plugin{
		name:    name,
		version: version,
		methods: map[string]AccessMethod{},

		downloaders:  map[string]Downloader{},
		downmappings: registry.NewRegistry[Downloader, DownloaderKey](),

		uploaders:  map[string]Uploader{},
		upmappings: registry.NewRegistry[Uploader, UploaderKey](),

		accessScheme:   runtime.MustNewDefaultScheme(&rt, &runtime.UnstructuredVersionedTypedObject{}, false, nil),
		uploaderScheme: runtime.MustNewDefaultScheme(&rt, &runtime.UnstructuredVersionedTypedObject{}, false, nil),

		actions: map[string]Action{},

		descriptor: descriptor.Descriptor{
			Version:       descriptor.VERSION,
			PluginName:    name,
			PluginVersion: version,
		},
	}
}

func (p *plugin) Name() string {
	return p.name
}

func (p *plugin) Version() string {
	return p.version
}

func (p *plugin) Descriptor() descriptor.Descriptor {
	if p.tweaker != nil {
		return p.tweaker(p.descriptor)
	}
	return p.descriptor
}

func (p *plugin) GetOptions() *Options {
	return &p.options
}

func (p *plugin) SetLong(s string) {
	p.descriptor.Long = s
}

func (p *plugin) SetShort(s string) {
	p.descriptor.Short = s
}

func (p *plugin) SetDescriptorTweaker(t func(descriptor descriptor.Descriptor) descriptor.Descriptor) {
	p.tweaker = t
}

func (p *plugin) SetConfigParser(config func(raw json.RawMessage) (interface{}, error)) {
	p.configParser = config
}

func (p *plugin) RegisterDownloader(arttype, mediatype string, hdlr Downloader) error {
	key := DownloaderKey{}.SetArtifact(arttype, mediatype)
	if !key.IsValid() {
		return errors.ErrInvalid("artifact context")
	}

	old := p.downloaders[hdlr.Name()]
	if old != nil && old != hdlr {
		return fmt.Errorf("downloader name %q already in use", hdlr.Name())
	}

	var desc *DownloaderDescriptor
	if old == nil {
		desc = &DownloaderDescriptor{
			Name:        hdlr.Name(),
			Description: hdlr.Description(),
			Constraints: []DownloaderKey{},
		}
		p.descriptor.Downloaders = append(p.descriptor.Downloaders, *desc)
		desc = &p.descriptor.Downloaders[len(p.descriptor.Downloaders)-1]
	} else {
		for i := range p.descriptor.Downloaders {
			if p.descriptor.Downloaders[i].Name == hdlr.Name() {
				desc = &p.descriptor.Downloaders[i]
			}
		}
	}

	cur := p.downmappings.GetHandler(key)
	if len(cur) > 0 && cur[0] != hdlr {
		return fmt.Errorf("downloader mapping key %q already in use", key)
	}
	if cur == nil {
		p.downmappings.Register(key, hdlr)
		desc.Constraints = append(desc.Constraints, DownloaderKey{ArtifactType: key.ArtifactType, MediaType: key.MediaType})
	}
	p.downloaders[hdlr.Name()] = hdlr
	return nil
}

func (p *plugin) GetDownloader(name string) Downloader {
	return p.downloaders[name]
}

func (p *plugin) GetDownloaderFor(arttype, mediatype string) Downloader {
	h := p.downmappings.LookupHandler(DownloaderKey{}.SetArtifact(arttype, mediatype))
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

func (p *plugin) RegisterRepositoryContextUploader(contexttype, repotype, arttype, mediatype string, u Uploader) error {
	if contexttype == "" || repotype == "" {
		return fmt.Errorf("repository context required")
	}
	return p.registerUploader(UploaderKey{}.SetArtifact(arttype, mediatype).SetRepo(contexttype, repotype), u)
}

func (p *plugin) RegisterUploader(arttype, mediatype string, u Uploader) error {
	return p.registerUploader(UploaderKey{}.SetArtifact(arttype, mediatype), u)
}

func (p *plugin) registerUploader(key UploaderKey, hdlr Uploader) error {
	if !key.RepositoryContext.IsValid() {
		return errors.ErrInvalid("repository context")
	}
	if !key.ArtifactContext.IsValid() {
		return errors.ErrInvalid("artifact context")
	}
	old := p.uploaders[hdlr.Name()]
	if old != nil && old != hdlr {
		return fmt.Errorf("uploader name %q already in use", hdlr.Name())
	}

	var desc *UploaderDescriptor
	if old == nil {
		desc = &UploaderDescriptor{
			Name:        hdlr.Name(),
			Description: hdlr.Description(),
			Constraints: []UploaderKey{},
		}
		p.descriptor.Uploaders = append(p.descriptor.Uploaders, *desc)
		desc = &p.descriptor.Uploaders[len(p.descriptor.Uploaders)-1]
	} else {
		for i := range p.descriptor.Uploaders {
			if p.descriptor.Uploaders[i].Name == hdlr.Name() {
				desc = &p.descriptor.Uploaders[i]
			}
		}
	}

	cur := p.upmappings.GetHandler(key)
	if len(cur) > 0 && cur[0] != hdlr {
		return fmt.Errorf("uploader mapping key %q already in use", key)
	}
	list := errors.ErrListf("uploader decoders")
	for n, d := range hdlr.Decoders() {
		list.Add(p.uploaderScheme.RegisterByDecoder(n, d))
	}
	if list.Len() > 0 {
		return list.Result()
	}
	if cur == nil {
		p.upmappings.Register(key, hdlr)
		desc.Constraints = append(desc.Constraints, key)
	}
	p.uploaders[hdlr.Name()] = hdlr
	return nil
}

func (p *plugin) GetUploader(name string) Uploader {
	return p.uploaders[name]
}

func (p *plugin) GetUploaderFor(arttype, mediatype string) Uploader {
	h := p.upmappings.LookupHandler(UploaderKey{}.SetArtifact(arttype, mediatype))
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

func (p *plugin) DecodeUploadTargetSpecification(data []byte) (UploadTargetSpec, error) {
	o, err := p.uploaderScheme.Decode(data, nil)
	if err != nil {
		return nil, err
	}
	return o.(UploadTargetSpec), nil
}

func (p *plugin) RegisterAccessMethod(m AccessMethod) error {
	if p.GetAccessMethod(m.Name(), m.Version()) != nil {
		n := m.Name()
		if m.Version() != "" {
			n += runtime.VersionSeparator + m.Version()
		}
		return errors.ErrAlreadyExists(errors.KIND_ACCESSMETHOD, n)
	}

	var optlist []CLIOption
	for _, o := range m.Options() {
		known := options.DefaultRegistry.GetOptionType(o.GetName())
		if known != nil {
			if o.ValueType() != known.ValueType() {
				return fmt.Errorf("option type %s[%s] conflicts with standard option type using value type %s", o.GetName(), o.ValueType(), known.ValueType())
			}
			optlist = append(optlist, CLIOption{
				Name: o.GetName(),
			})
		} else {
			optlist = append(optlist, CLIOption{
				Name:        o.GetName(),
				Type:        o.ValueType(),
				Description: o.GetDescriptionText(),
			})
		}
	}
	vers := m.Version()
	if vers == "" {
		meth := descriptor.AccessMethodDescriptor{
			Name:        m.Name(),
			Description: m.Description(),
			Format:      m.Format(),
		}
		p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
		p.accessScheme.RegisterByDecoder(m.Name(), m)
		p.methods[m.Name()] = m
		vers = "v1"
	}
	meth := descriptor.AccessMethodDescriptor{
		Name:        m.Name(),
		Version:     vers,
		Description: m.Description(),
		Format:      m.Format(),
		CLIOptions:  optlist,
	}
	p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
	p.accessScheme.RegisterByDecoder(m.Name()+"/"+vers, m)
	p.methods[m.Name()+"/"+vers] = m
	return nil
}

func (p *plugin) DecodeAccessSpecification(data []byte) (AccessSpec, error) {
	o, err := p.accessScheme.Decode(data, nil)
	if err != nil {
		return nil, err
	}
	return o.(AccessSpec), nil
}

func (p *plugin) GetAccessMethod(name string, version string) AccessMethod {
	n := name
	if version != "" {
		n += "/" + version
	}
	return p.methods[n]
}

func (p *plugin) RegisterAction(a Action) error {
	if p.GetAction(a.Name()) != nil {
		return errors.ErrAlreadyExists("action", a.Name())
	}
	vers := action.SupportedActionVersions(a.Name())
	if len(vers) == 0 {
		return errors.ErrNotSupported("action", a.Name())
	}

	act := descriptor.ActionDescriptor{
		Name:             a.Name(),
		Versions:         vers,
		Description:      a.Description(),
		DefaultSelectors: a.DefaultSelectors(),
		ConsumerType:     a.ConsumerType(),
	}
	p.descriptor.Actions = append(p.descriptor.Actions, act)
	p.actions[a.Name()] = a
	return nil
}

func (p *plugin) DecodeAction(data []byte) (ActionSpec, error) {
	return action.DecodeActionSpec(data)
}

func (p *plugin) GetAction(name string) Action {
	return p.actions[name]
}

func (p *plugin) GetConfig() (interface{}, error) {
	if len(p.options.Config) == 0 {
		return nil, nil
	}
	if p.configParser == nil {
		var cfg interface{}
		if err := json.Unmarshal(p.options.Config, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	return p.configParser(p.options.Config)
}
