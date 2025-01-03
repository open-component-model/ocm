package accessors

import (
	"ocm.software/ocm/api/ocm/compdesc/equivalent"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/utils/runtime"
)

// ElementListAccessor provides generic access to list of elements.
type ElementListAccessor interface {
	Len() int
	Get(i int) Element
}

// ElementMeta describes the access to common element meta data attributes.
type ElementMeta interface {
	GetName() string
	GetVersion() string
	GetExtraIdentity() v1.Identity
	GetLabels() v1.Labels
	GetIdentity(accessor ElementListAccessor) v1.Identity
	GetIdentityDigest(accessor ElementListAccessor) []byte

	GetRawIdentity() v1.Identity
	GetMatchBaseIdentity() v1.Identity

	GetMeta() ElementMeta // ElementMeta is again a Meta provider

	SetLabels(labels []v1.Label)
	SetExtraIdentity(identity v1.Identity)
}

// ElementMetaProvider just provides access to element meta data
// of an element.
type ElementMetaProvider interface {
	GetMeta() ElementMeta
}

// Element represents a generic element with meta information.
// It also allows to check for equivalence of elements of the same kind.
type Element interface {
	ElementMetaProvider
	Equivalent(Element) equivalent.EqualState
}

// AccessSpec is the minimal interface  for access spec attributes.
type AccessSpec interface {
	runtime.VersionedTypedObject
}

// ArtifactAccessor provides access to generic artifact information of an element.
type ArtifactAccessor interface {
	Element
	GetType() string
	GetAccess() AccessSpec
}

// ResourceAccessor provides access to resource attribute.
type ResourceAccessor interface {
	ArtifactAccessor
	GetRelation() v1.ResourceRelation
	GetDigest() *v1.DigestSpec
}

// SourceAccessor provides access to source attribute.
type SourceAccessor interface {
	ArtifactAccessor
}

// ReferenceAccessor provides access to source attribute.
type ReferenceAccessor interface {
	Element
	GetComponentName() string
}
