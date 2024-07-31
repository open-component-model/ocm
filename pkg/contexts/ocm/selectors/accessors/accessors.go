package accessors

import (
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// ElementListAccessor provides generic access to list of elements.
type ElementListAccessor interface {
	Len() int
	Get(i int) ElementMetaAccessor
}

// ElementMeta describes the access to common element meta data attributes.
type ElementMeta interface {
	GetName() string
	GetVersion() string
	GetExtraIdentity() v1.Identity
	GetLabels() v1.Labels
	GetIdentityForContext(accessor ElementListAccessor) v1.Identity
}

// ElementMetaAccessor provides generic access an elements meta information.
type ElementMetaAccessor interface {
	GetMeta() ElementMeta
}

// AccessSpec is the minimal interface  for access spec attributes.
type AccessSpec interface {
	runtime.VersionedTypedObject
}

// ArtifactAccessor provides access to generic artifact information of an element.
type ArtifactAccessor interface {
	ElementMetaAccessor
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
	ElementMetaAccessor
	GetComponentName() string
}
