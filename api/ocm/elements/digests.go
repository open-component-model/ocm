package elements

import (
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
)

type ResourceReferenceOption interface {
	ResourceMetaOption
	ReferenceOption
}

////////////////////////////////////////////////////////////////////////////////

type digest metav1.DigestSpec

func (o *digest) ApplyToReference(m *compdesc.ComponentReference) error {
	if !(*metav1.DigestSpec)(o).IsNone() {
		m.Digest = (*metav1.DigestSpec)(o).Copy()
	}
	return nil
}

func (o *digest) ApplyToResourceMeta(m *compdesc.ResourceMeta) error {
	if !(*metav1.DigestSpec)(o).IsNone() {
		m.Digest = (*metav1.DigestSpec)(o).Copy()
	}
	return nil
}

// WithDigest sets digest information.
// at least one value should be set.
func WithDigest(algo, norm, value string) ResourceReferenceOption {
	return &digest{HashAlgorithm: algo, NormalisationAlgorithm: norm, Value: value}
}
