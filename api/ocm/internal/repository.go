package internal

import (
	"io"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/refhints"
	"ocm.software/ocm/api/ocm/selectors/refsel"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/ocm/selectors/srcsel"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

type ReadOnlyFeature interface {
	IsReadOnly() bool
	// SetReadOnly is used to set the element into readonly mode.
	// Once enabled it cannot be reverted. An underlying object, for
	// example a CTF might be in readonly mode, forced by filesystem
	// permissions. Such elements cannot be set into write mode again.
	// Therefore, generally only one direction is possible.
	SetReadOnly()
}

type RepositoryImpl interface {
	GetContext() Context

	GetSpecification() RepositorySpec
	ComponentLister() ComponentLister

	ExistsComponentVersion(name string, version string) (bool, error)
	LookupComponentVersion(name string, version string) (ComponentVersionAccess, error)
	LookupComponent(name string) (ComponentAccess, error)

	io.Closer
	ReadOnlyFeature
}

type Repository interface {
	resource.ResourceView[Repository]
	RepositoryImpl

	NewComponentVersion(comp, version string, overrides ...bool) (ComponentVersionAccess, error)
	AddComponentVersion(cv ComponentVersionAccess, overrides ...bool) error
}

// ConsumerIdentityProvider is an interface for object requiring
// credentials, which want to expose the ConsumerId they are
// usingto request implicit credentials.
type ConsumerIdentityProvider = credentials.ConsumerIdentityProvider

type (
	DataAccess = blobaccess.DataAccess
	BlobAccess = blobaccess.BlobAccess
	MimeType   = blobaccess.MimeType
)

type ComponentAccess interface {
	resource.ResourceView[ComponentAccess]

	GetContext() Context
	GetName() string

	ListVersions() ([]string, error)
	LookupVersion(version string) (ComponentVersionAccess, error)
	HasVersion(vers string) (bool, error)
	NewVersion(version string, overrides ...bool) (ComponentVersionAccess, error)
	AddVersion(cv ComponentVersionAccess, overrides ...bool) error
	AddVersionOpt(cv ComponentVersionAccess, opts ...AddVersionOption) error

	io.Closer
}

// AccessProvider assembled methods provided
// and used for access methods.
// It is provided for resources in a component version
// with the base support implementation in package cpi.
// But it can also be provided by resource provisioners
// used to feed the ComponentVersionAccess.SetResourceByAccess
// or the ComponentVersionAccessSetSourceByAccess
// method.
type AccessProvider interface {
	GetOCMContext() Context

	// ReferenceHintForAccess is the default hint representation provided
	// by an access method. This is typically a single hint.
	// Additionally, the artifact meta data might offer explicit hints.
	ReferenceHintForAccess() refhints.ReferenceHints

	Access() (AccessSpec, error)
	AccessMethod() (AccessMethod, error)

	GlobalAccess() AccessSpec

	blobaccess.BlobAccessProvider
}

type ArtifactAccess[M any] interface {
	Meta() *M
	GetComponentVersion() (ComponentVersionAccess, error)
	AccessProvider

	// GetReferenceHint provides the effective hints
	// for an artifact. It is composed of the
	// hints explicitly given by the artifact metadata
	// and optional hints provided by the access method.
	GetReferenceHints() refhints.ReferenceHints
}

type (
	ResourceMeta   = compdesc.ResourceMeta
	ResourceAccess = ArtifactAccess[ResourceMeta]
)

type (
	SourceMeta   = compdesc.SourceMeta
	SourceAccess = ArtifactAccess[SourceMeta]
)

type ComponentReference = compdesc.Reference

type ComponentVersionAccess interface {
	resource.ResourceView[ComponentVersionAccess]
	common.VersionedElement
	io.Closer
	ReadOnlyFeature

	GetContext() Context
	Repository() Repository
	GetDescriptor() *compdesc.ComponentDescriptor

	DiscardChanges()
	IsPersistent() bool

	GetProvider() *compdesc.Provider
	SetProvider(provider *compdesc.Provider) error

	GetResource(meta metav1.Identity) (ResourceAccess, error)
	GetResourceIndex(meta metav1.Identity) int
	GetResourceByIndex(i int) (ResourceAccess, error)
	GetResources() []ResourceAccess
	SelectResources(sel ...rscsel.Selector) ([]ResourceAccess, error)

	SetResource(*ResourceMeta, compdesc.AccessSpec, ...ModificationOption) error
	SetResourceByAccess(art ResourceAccess, modopts ...BlobModificationOption) error

	GetSource(meta metav1.Identity) (SourceAccess, error)
	GetSourceIndex(meta metav1.Identity) int
	GetSourceByIndex(i int) (SourceAccess, error)
	GetSources() []SourceAccess
	SelectSources(sel ...srcsel.Selector) ([]SourceAccess, error)

	// SetSource updates or sets anew source. The options only use the
	// target options. All other options are ignored.
	SetSource(*SourceMeta, compdesc.AccessSpec, ...TargetElementOption) error
	// SetSourceByAccess updates or sets anew source. The options only use the
	// target options. All other options are ignored.
	SetSourceByAccess(art SourceAccess, opts ...TargetElementOption) error

	GetReference(meta metav1.Identity) (ComponentReference, error)
	GetReferenceIndex(meta metav1.Identity) int
	GetReferenceByIndex(i int) (ComponentReference, error)
	GetReferences() []ComponentReference
	SelectReferences(sel ...refsel.Selector) ([]ComponentReference, error)

	// SetReference adds or updates a reference. By default, it does not allow for
	// signature relevant changes. If such operations should be possible
	// the option ModifyElement() has to be passed as option.
	SetReference(ref *ComponentReference, opts ...ElementModificationOption) error

	// AddBlob adds a local blob and returns an appropriate local access spec.
	AddBlob(blob BlobAccess, artType string, hints []refhints.ReferenceHint, global AccessSpec, opts ...BlobUploadOption) (AccessSpec, error)

	// AdjustResourceAccess is used to modify the access spec. The old and new one MUST refer to the same content.
	AdjustResourceAccess(meta *ResourceMeta, acc compdesc.AccessSpec, opts ...ModificationOption) error
	SetResourceBlob(meta *ResourceMeta, blob BlobAccess, hints []refhints.ReferenceHint, global AccessSpec, opts ...BlobModificationOption) error
	AdjustSourceAccess(meta *SourceMeta, acc compdesc.AccessSpec) error
	// SetSourceBlob updates or sets anew source. The options only use the
	// target options. All other options are ignored.
	SetSourceBlob(meta *SourceMeta, blob BlobAccess, hints []refhints.ReferenceHint, global AccessSpec, opts ...TargetElementOption) error

	// AccessMethod provides an access method implementation for
	// an access spec. This might be a repository local implementation
	// or a global one. It might be called by the AccessSpec method
	// to map itself to a local implementation or called directly.
	// If called it should forward the call to the AccessSpec
	// if and only if this specs NOT states to be local IsLocal()==false
	// If the spec states to be local, the repository is responsible for
	// providing a local implementation or return nil if this is
	// not supported by the actual repository type.
	AccessMethod(AccessSpec) (AccessMethod, error)

	// Update adds the version with all changes to the component instance it has been created for.
	Update() error

	// Execute executes a function on a valid and locked component version reference.
	// If it returns an error this error is forwarded.
	Execute(func() error) error
}

// ComponentLister provides the optional repository list functionality of
// a repository.
type ComponentLister interface {
	// NumComponents returns the number of components found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	NumComponents(prefix string) (int, error)

	// GetNamespaces returns the name of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	GetComponents(prefix string, closure bool) ([]string, error)
}
