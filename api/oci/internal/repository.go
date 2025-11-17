package internal

import (
	"io"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt/resource"
)

type RepositoryImpl interface {
	GetSpecification() RepositorySpec
	GetContext() Context

	NamespaceLister() NamespaceLister
	ExistsArtifact(name string, ref string) (bool, error)
	LookupArtifact(name string, ref string) (ArtifactAccess, error)
	LookupNamespace(name string) (NamespaceAccess, error)

	io.Closer
}

type Repository interface {
	resource.ResourceView[Repository]

	RepositoryImpl
}

// ConsumerIdentityProvider is an optional interface for repositories
// to tell about their credential requests.
type ConsumerIdentityProvider = credentials.ConsumerIdentityProvider

type RepositorySource interface {
	GetRepository() Repository
}

type (
	BlobAccess = blobaccess.BlobAccess
	DataAccess = blobaccess.DataAccess
)

type BlobSource interface {
	GetBlobData(digest digest.Digest) (int64, DataAccess, error)
}

type BlobSink interface {
	AddBlob(BlobAccess) error
}

type ArtifactSink interface {
	AddBlob(BlobAccess) error
	AddArtifact(a Artifact, tags ...string) (BlobAccess, error)
	AddTags(digest digest.Digest, tags ...string) error
}

type ArtifactSource interface {
	GetArtifact(version string) (ArtifactAccess, error)
	GetBlobData(digest digest.Digest) (int64, DataAccess, error)
}

type NamespaceAccessImpl interface {
	ArtifactSource
	ArtifactSink
	GetNamespace() string
	ListTags() ([]string, error)

	HasArtifact(vers string) (bool, error)

	NewArtifact(...Artifact) (ArtifactAccess, error)
	io.Closer
}

type NamespaceAccess interface {
	resource.ResourceView[NamespaceAccess]

	NamespaceAccessImpl
}

type Artifact artdesc.ArtifactDescriptor

type ArtifactAccessImpl interface {
	Artifact
	BlobSource
	BlobSink

	GetDescriptor() *artdesc.Artifact
	GetBlob(digest digest.Digest) (BlobAccess, error)

	GetArtifact(digest digest.Digest) (ArtifactAccess, error)
	AddBlob(BlobAccess) error

	AddArtifact(Artifact, *artdesc.Platform) (BlobAccess, error)
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)

	NewArtifact(...Artifact) (ArtifactAccess, error)

	io.Closer
}

type ArtifactAccessSlaves interface {
	ManifestAccess() ManifestAccess
	IndexAccess() IndexAccess
}

type ArtifactAccess interface {
	resource.ResourceView[ArtifactAccess]

	ArtifactAccessImpl
	ArtifactAccessSlaves
}

type ManifestAccess interface {
	Artifact

	GetDescriptor() *artdesc.Manifest
	GetBlobDescriptor(digest digest.Digest) *artdesc.Descriptor
	GetConfigBlob() (BlobAccess, error)
	GetBlob(digest digest.Digest) (BlobAccess, error)

	AddBlob(BlobAccess) error
	AddLayer(BlobAccess, *artdesc.Descriptor) (int, error)
	Modify(func(manifest *artdesc.Manifest) error) error
	SetConfigBlob(blob BlobAccess, d *artdesc.Descriptor) error
}

type IndexAccess interface {
	Artifact

	GetDescriptor() *artdesc.Index
	GetBlobDescriptor(digest digest.Digest) *artdesc.Descriptor
	GetBlob(digest digest.Digest) (BlobAccess, error)

	GetArtifact(digest digest.Digest) (ArtifactAccess, error)
	/*
		GetIndex(digest digest.Digest) (IndexAccess, error)
		GetManifest(digest digest.Digest) (ManifestAccess, error)
	*/

	AddBlob(BlobAccess) error
	AddArtifact(Artifact, *artdesc.Platform) (BlobAccess, error)
}

// NamespaceLister provides the optional repository list functionality of
// a repository.
type NamespaceLister interface {
	// NumNamespaces returns the number of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	NumNamespaces(prefix string) (int, error)

	// GetNamespaces returns the name of namespaces found for a prefix
	// If the given prefix does not end with a /, a repository with the
	// prefix name is included
	GetNamespaces(prefix string, closure bool) ([]string, error)
}
