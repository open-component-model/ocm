// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ppi

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/slices"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
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
	uploaderScheme runtime.Scheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]

	methods      map[string]AccessMethod
	accessScheme runtime.Scheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]

	actions       map[string]Action
	mergehandlers map[string]ValueMergeHandler
	mergespecs    map[string]*descriptor.LabelMergeSpecification

	valuesets map[string]map[string]ValueSet
	setScheme map[string]runtime.Scheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]

	configParser func(message json.RawMessage) (interface{}, error)
}

func NewPlugin(name string, version string) Plugin {
	return &plugin{
		name:    name,
		version: version,
		methods: map[string]AccessMethod{},

		downloaders:  map[string]Downloader{},
		downmappings: registry.NewRegistry[Downloader, DownloaderKey](),

		uploaders:  map[string]Uploader{},
		upmappings: registry.NewRegistry[Uploader, UploaderKey](),

		accessScheme:   runtime.MustNewDefaultScheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]](&runtime.UnstructuredVersionedTypedObject{}, false, nil),
		uploaderScheme: runtime.MustNewDefaultScheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]](&runtime.UnstructuredVersionedTypedObject{}, false, nil),

		actions:       map[string]Action{},
		mergehandlers: map[string]ValueMergeHandler{},
		mergespecs:    map[string]*descriptor.LabelMergeSpecification{},

		valuesets: map[string]map[string]ValueSet{},
		setScheme: map[string]runtime.Scheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]]{},

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
		schema := ""
		if len(hdlr.ConfigSchema()) > 0 {
			schema = string(hdlr.ConfigSchema())
		}
		desc = &DownloaderDescriptor{
			Name:         hdlr.Name(),
			Description:  hdlr.Description(),
			Constraints:  []DownloaderKey{},
			ConfigScheme: schema,
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
	return o, nil
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
	_, idp := m.(ContentVersionIdentityProvider)
	vers := m.Version()
	if vers == "" {
		meth := descriptor.AccessMethodDescriptor{
			ValueSetDefinition: descriptor.ValueSetDefinition{
				Name:        m.Name(),
				Description: m.Description(),
				Format:      m.Format(),
			},
			SupportContentIdentity: idp,
		}
		p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
		p.accessScheme.RegisterByDecoder(m.Name(), m)
		p.methods[m.Name()] = m
		vers = "v1"
	}
	meth := descriptor.AccessMethodDescriptor{
		ValueSetDefinition: descriptor.ValueSetDefinition{
			Name:        m.Name(),
			Version:     vers,
			Description: m.Description(),
			Format:      m.Format(),
			CLIOptions:  optlist,
		},
		SupportContentIdentity: idp,
	}
	p.descriptor.AccessMethods = append(p.descriptor.AccessMethods, meth)
	p.accessScheme.RegisterByDecoder(m.Name()+"/"+vers, m)
	p.methods[m.Name()+"/"+vers] = m
	return nil
}

func (p *plugin) DecodeAccessSpecification(data []byte) (AccessSpec, error) {
	return p.accessScheme.Decode(data, nil)
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
	vers := action.DefaultRegistry().SupportedActionVersions(a.Name())
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
	return action.DefaultRegistry().DecodeActionSpec(data, runtime.DefaultJSONEncoding)
}

func (p *plugin) GetAction(name string) Action {
	return p.actions[name]
}

func (p *plugin) RegisterValueMergeHandler(a ValueMergeHandler) error {
	if p.GetValueMergeHandler(a.Name()) != nil {
		return errors.ErrAlreadyExists("value mergehandler", a.Name())
	}

	hd := descriptor.ValueMergeHandlerDescriptor{
		Name:        a.Name(),
		Description: a.Description(),
	}
	p.descriptor.ValueMergeHandlers = append(p.descriptor.ValueMergeHandlers, hd)
	p.mergehandlers[a.Name()] = a
	return nil
}

func (p *plugin) GetValueMergeHandler(name string) ValueMergeHandler {
	return p.mergehandlers[name]
}

func (p *plugin) RegisterLabelMergeSpecification(name, version string, spec *metav1.MergeAlgorithmSpecification, desc string) error {
	e := descriptor.LabelMergeSpecification{
		Name:                        name,
		Version:                     version,
		Description:                 desc,
		MergeAlgorithmSpecification: *spec,
	}

	if p.GetLabelMergeSpecification(e.GetName()) != nil {
		return errors.ErrAlreadyExists("label merge spec", e.GetName())
	}

	p.descriptor.LabelMergeSpecifications = append(p.descriptor.LabelMergeSpecifications, e)
	p.mergespecs[e.GetName()] = &e
	return nil
}

func (p *plugin) GetLabelMergeSpecification(id string) *descriptor.LabelMergeSpecification {
	return p.mergespecs[id]
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

func (p *plugin) DecodeValueSet(purpose string, data []byte) (runtime.TypedObject, error) {
	schemes := p.setScheme[purpose]
	if schemes == nil {
		return nil, errors.ErrUnknown(descriptor.KIND_PURPOSE)
	}
	return schemes.Decode(data, nil)
}

func (p *plugin) GetValueSet(purpose string, name string, version string) ValueSet {
	n := name
	if version != "" {
		n += "/" + version
	}
	set := p.valuesets[purpose]
	if set == nil {
		return nil
	}
	return set[n]
}

func (p *plugin) RegisterValueSet(s ValueSet) error {
	n := s.Name()
	if s.Version() != "" {
		n += runtime.VersionSeparator + s.Version()
	}
	for _, pp := range s.Purposes() {
		if p.GetValueSet(pp, s.Name(), s.Version()) != nil {
			return errors.ErrAlreadyExists(descriptor.KIND_VALUESET, n)
		}
	}

	var optlist []CLIOption
	for _, o := range s.Options() {
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
	vers := s.Version()
	if vers == "" {
		set := descriptor.ValueSetDescriptor{
			ValueSetDefinition: descriptor.ValueSetDefinition{
				Name:        s.Name(),
				Description: s.Description(),
				Format:      s.Format(),
			},
			Purposes: slices.Clone(s.Purposes()),
		}
		p.descriptor.ValueSets = append(p.descriptor.ValueSets, set)
		for _, pp := range s.Purposes() {
			schemes := p.setScheme[pp]
			if schemes == nil {
				schemes = runtime.MustNewDefaultScheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]](&runtime.UnstructuredVersionedTypedObject{}, false, nil)
				p.setScheme[pp] = schemes
			}
			schemes.RegisterByDecoder(s.Name(), s)
			sets := p.valuesets[pp]
			if sets == nil {
				sets = map[string]ValueSet{}
				p.valuesets[pp] = sets
			}
			sets[s.Name()] = s
		}
		vers = "v1"
	}
	set := descriptor.ValueSetDescriptor{
		ValueSetDefinition: descriptor.ValueSetDefinition{
			Name:        s.Name(),
			Version:     vers,
			Description: s.Description(),
			Format:      s.Format(),
			CLIOptions:  optlist,
		},
		Purposes: slices.Clone(s.Purposes()),
	}
	p.descriptor.ValueSets = append(p.descriptor.ValueSets, set)
	for _, pp := range s.Purposes() {
		schemes := p.setScheme[pp]
		if schemes == nil {
			schemes = runtime.MustNewDefaultScheme[runtime.TypedObject, runtime.TypedObjectDecoder[runtime.TypedObject]](&runtime.UnstructuredVersionedTypedObject{}, false, nil)
			p.setScheme[pp] = schemes
		}
		schemes.RegisterByDecoder(s.Name()+"/"+vers, s)
		sets := p.valuesets[pp]
		if sets == nil {
			sets = map[string]ValueSet{}
			p.valuesets[pp] = sets
		}
		sets[s.Name()+"/"+vers] = s
	}
	return nil
}
