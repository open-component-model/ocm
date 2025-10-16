package support

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/refmgmt"
)

// BlobProvider manages the technical access to blobs.
type BlobProvider interface {
	refmgmt.Allocatable
	cpi.BlobSource
	cpi.BlobSink
}

// NamespaceContainer is the interface used by subsequent access objects
// to access the base implementation.
type NamespaceContainer interface {
	SetImplementation(impl NamespaceAccessImpl)

	IsReadOnly() bool
	// IsClosed() bool

	cpi.BlobSource
	cpi.BlobSink

	Close() error

	// GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor

	GetArtifact(i NamespaceAccessImpl, vers string) (cpi.ArtifactAccess, error)
	NewArtifact(i NamespaceAccessImpl, arts ...cpi.Artifact) (cpi.ArtifactAccess, error)

	AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error)

	AddTags(digest digest.Digest, tags ...string) error
	ListTags() ([]string, error)
	HasArtifact(vers string) (bool, error)
}

////////////////////////////////////////////////////////////////////////////////

type NamespaceAccessImpl interface {
	cpi.NamespaceAccessImpl

	// GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor
	IsReadOnly() bool

	WithContainer(container NamespaceContainer) NamespaceAccessImpl
}

type namespaceAccessImpl struct {
	*cpi.NamespaceAccessImplBase
	NamespaceContainer // inherit as many as possible methods for cpi.NamespaceAccessImpl
}

var _ NamespaceAccessImpl = (*namespaceAccessImpl)(nil)

func NewNamespaceAccessImpl(namespace string, c NamespaceContainer, repo cpi.RepositoryViewManager) (NamespaceAccessImpl, error) {
	base, err := cpi.NewNamespaceAccessImplBase(namespace, repo)
	if err != nil {
		return nil, err
	}
	impl := &namespaceAccessImpl{
		NamespaceAccessImplBase: base,
		NamespaceContainer:      c,
	}

	c.SetImplementation(impl)
	return impl, nil
}

func (n *namespaceAccessImpl) Close() error {
	return accessio.Close(n.NamespaceAccessImplBase, n.NamespaceContainer)
}

func NewNamespaceAccess(namespace string, c NamespaceContainer, repo cpi.RepositoryViewManager, kind ...string) (cpi.NamespaceAccess, error) {
	impl, err := NewNamespaceAccessImpl(namespace, c, repo)
	if err != nil {
		return nil, err
	}
	return cpi.NewNamespaceAccess(impl, kind...), nil
}

func GetArtifactSetContainer(i cpi.NamespaceAccessImpl) (NamespaceContainer, error) {
	if c, ok := i.(*namespaceAccessImpl); ok {
		return c.NamespaceContainer, nil
	}
	return nil, errors.ErrNotSupported()
}

func (i *namespaceAccessImpl) WithContainer(c NamespaceContainer) NamespaceAccessImpl {
	return &namespaceAccessImpl{
		NamespaceAccessImplBase: i.NamespaceAccessImplBase,
		NamespaceContainer:      c,
	}
}

func (i *namespaceAccessImpl) GetArtifact(vers string) (cpi.ArtifactAccess, error) {
	return i.NamespaceContainer.GetArtifact(i, vers)
}

func (i *namespaceAccessImpl) AddArtifact(artifact cpi.Artifact, tags ...string) (access blobaccess.BlobAccess, err error) {
	return i.NamespaceContainer.AddArtifact(artifact, tags...)
}

func (i *namespaceAccessImpl) NewArtifact(arts ...cpi.Artifact) (cpi.ArtifactAccess, error) {
	return i.NamespaceContainer.NewArtifact(i, arts...)
}
