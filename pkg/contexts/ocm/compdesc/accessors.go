package compdesc

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/equivalent"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors/accessors"
)

// NameAccessor describes a accessor for a named object.
type NameAccessor interface {
	// GetName returns the name of the object.
	GetName() string
	// SetName sets the name of the object.
	SetName(name string)
}

// VersionAccessor describes a accessor for a versioned object.
type VersionAccessor interface {
	// GetVersion returns the version of the object.
	GetVersion() string
	// SetVersion sets the version of the object.
	SetVersion(version string)
}

// LabelsAccessor describes a accessor for a labeled object.
type LabelsAccessor interface {
	// GetLabels returns the labels of the object.
	GetLabels() metav1.Labels
	// SetLabels sets the labels of the object.
	SetLabels(labels []metav1.Label)
}

// ObjectMetaAccessor describes a accessor for named and versioned object.
type ObjectMetaAccessor interface {
	NameAccessor
	VersionAccessor
	LabelsAccessor
}

// ElementMetaAccessor provides generic access an elements meta information.
type ElementMetaAccessor interface {
	ElementMetaProvider
	Equivalent(ElementMetaAccessor) equivalent.EqualState
}

// ElementAccessor provides generic access to list of elements.
type ElementAccessor interface {
	Len() int
	Get(i int) ElementMetaAccessor
}

type ElementMetaProvider interface {
	GetMeta() *ElementMeta
}

// ElementArtifactAccessor provides access to generic artifact information of an element.
type ElementArtifactAccessor interface {
	ElementMetaAccessor
	GetType() string
	GetAccess() AccessSpec
	SetAccess(a AccessSpec)
}

type ElementDigestAccessor interface {
	GetDigest() *metav1.DigestSpec
	SetDigest(*metav1.DigestSpec)
}

// ArtifactAccessor provides generic access to list of artifacts.
// There are resources or sources.
type ArtifactAccessor interface {
	ElementAccessor
	GetArtifact(i int) ElementArtifactAccessor
}

// AccessSpec is an abstract specification of an access method
// The outbound object is typicall a runtime.UnstructuredTypedObject.
// Inbound any serializable AccessSpec implementation is possible.
type AccessSpec = accessors.AccessSpec

// AccessProvider provides access to an access specification of elements.
type AccessProvider interface {
	GetAccess() AccessSpec
}
