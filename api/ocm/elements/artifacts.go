package elements

import (
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/refhints"
)

type ArtifactOption interface {
	ResourceMetaOption
	SourceMetaOption
}

type artifactAccessor interface {
	compdesc.ReferenceHintSink
}

type artifactOption struct {
	apply func(meta artifactAccessor) error
}

type artifactOptionI interface {
	apply(sink artifactAccessor) error
}

func newArtifactOption[T artifactOptionI](e T) ArtifactOption {
	return artifactOption{e.apply}
}

func (o artifactOption) ApplyToResourceMeta(m *compdesc.ResourceMeta) error {
	return o.apply(m)
}

func (o artifactOption) ApplyToSourceMeta(m *compdesc.SourceMeta) error {
	return o.apply(m)
}

////////////////////////////////////////////////////////////////////////////////

type refhint string

func (o refhint) apply(m artifactAccessor) error {
	m.SetReferenceHints(refhints.ParseHints(string(o)))
	return nil
}

// WithHint sets a serialized list of hints for the artifact metadata.
func WithHint(v string) ArtifactOption {
	return newArtifactOption(refhint(v))
}
