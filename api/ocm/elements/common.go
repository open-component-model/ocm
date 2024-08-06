package elements

import (
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

type CommonOption interface {
	ResourceMetaOption
	SourceMetaOption
	ReferenceOption
}

type commonOption struct {
	apply func(meta *compdesc.ElementMeta) error
}

type commonOptionI interface {
	apply(*compdesc.ElementMeta) error
}

func newCommonOption[T commonOptionI](e T) CommonOption {
	return commonOption{e.apply}
}

func (o commonOption) ApplyToResourceMeta(m *compdesc.ResourceMeta) error {
	return o.apply(&m.ElementMeta)
}

func (o commonOption) ApplyToSourceMeta(m *compdesc.SourceMeta) error {
	return o.apply(&m.ElementMeta)
}

func (o commonOption) ApplyToReference(m *compdesc.Reference) error {
	return o.apply(&m.ElementMeta)
}

////////////////////////////////////////////////////////////////////////////////

type version string

func (o version) apply(m *compdesc.ElementMeta) error {
	m.Version = string(o)
	return nil
}

// WithVersion sets the version of the element.
func WithVersion(v string) CommonOption {
	return newCommonOption(version(v))
}

////////////////////////////////////////////////////////////////////////////////

type extraIdentity struct {
	id metav1.Identity
}

func (o *extraIdentity) apply(m *compdesc.ElementMeta) error {
	if m.ExtraIdentity == nil {
		m.ExtraIdentity = o.id.Copy()
	} else {
		for n, v := range o.id {
			m.ExtraIdentity.Set(n, v)
		}
	}
	return nil
}

// WithExtraIdentity adds extra identity properties.
func WithExtraIdentity(extras ...string) CommonOption {
	return newCommonOption(&extraIdentity{compdesc.NewExtraIdentity(extras...)})
}

////////////////////////////////////////////////////////////////////////////////

type label struct {
	name  string
	value interface{}
	opts  []metav1.LabelOption
}

func (o *label) apply(m *compdesc.ElementMeta) error {
	return m.Labels.Set(o.name, o.value, o.opts...)
}

func WithLabel(name string, value interface{}, opts ...metav1.LabelOption) CommonOption {
	return newCommonOption(&label{name, value, opts})
}
