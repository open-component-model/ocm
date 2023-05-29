// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package support

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

// BlobProvider manages the technical access to blobs.
type BlobProvider interface {
	accessio.Allocatable
	cpi.BlobSource
	cpi.BlobSink
}

// ArtifactSetContainer is the interface used by subsequent access objects
// to access the base implementation.
type ArtifactSetContainer interface {
	GetNamespace() string

	IsReadOnly() bool
	// IsClosed() bool

	cpi.BlobSource
	cpi.BlobSink

	Close() error

	// GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor

	GetArtifact(i ArtifactSetImpl, vers string) (cpi.ArtifactAccess, error)
	NewArtifact(i ArtifactSetImpl, arts ...*artdesc.Artifact) (cpi.ArtifactAccess, error)

	AddArtifact(artifact cpi.Artifact, tags ...string) (access accessio.BlobAccess, err error)

	AddTags(digest digest.Digest, tags ...string) error
	ListTags() ([]string, error)
	HasArtifact(vers string) (bool, error)
}

////////////////////////////////////////////////////////////////////////////////

type ArtifactSetImpl interface {
	cpi.NamespaceAccessImpl

	View(main ...bool) (cpi.NamespaceAccess, error)

	// GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor
	IsReadOnly() bool

	WithContainer(container ArtifactSetContainer) ArtifactSetImpl
}

type artifactSetImpl struct {
	refs                 cpi.NamespaceAccessViewManager
	ArtifactSetContainer // inherit as many as possible methods for cpi.NamespaceAccessImpl
}

var _ ArtifactSetImpl = (*artifactSetImpl)(nil)

func NewArtifactSetContainerImpl(c ArtifactSetContainer) ArtifactSetImpl {
	return &artifactSetImpl{
		ArtifactSetContainer: c,
	}
}

func NewArtifactSet(c ArtifactSetContainer, kind ...string) cpi.NamespaceAccess {
	return cpi.NewNamespaceAccess(NewArtifactSetContainerImpl(c), kind...)
}

func GetArtifactSetContainer(i cpi.NamespaceAccessImpl) (ArtifactSetContainer, error) {
	if c, ok := i.(*artifactSetImpl); ok {
		return c.ArtifactSetContainer, nil
	}
	return nil, errors.ErrNotSupported()
}

func (i *artifactSetImpl) SetViewManager(m cpi.NamespaceAccessViewManager) {
	i.refs = m
}

func (i *artifactSetImpl) WithContainer(c ArtifactSetContainer) ArtifactSetImpl {
	return &artifactSetImpl{
		refs:                 i.refs,
		ArtifactSetContainer: c,
	}
}

func (i *artifactSetImpl) View(main ...bool) (cpi.NamespaceAccess, error) {
	return i.refs.View(main...)
}

func (i *artifactSetImpl) GetArtifact(vers string) (cpi.ArtifactAccess, error) {
	return i.ArtifactSetContainer.GetArtifact(i, vers)
}

func (i *artifactSetImpl) AddArtifact(artifact cpi.Artifact, tags ...string) (access accessio.BlobAccess, err error) {
	return i.ArtifactSetContainer.AddArtifact(artifact, tags...)
}

func (i *artifactSetImpl) NewArtifact(arts ...*artdesc.Artifact) (cpi.ArtifactAccess, error) {
	return i.ArtifactSetContainer.NewArtifact(i, arts...)
}
