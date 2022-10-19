// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf

import (
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi/support"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/index"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Namespace struct {
	view accessio.CloserView
	*NamespaceContainer
}

// implemented by view
// the rest is directly taken from the artefact set implementation

func (s *Namespace) Close() error {
	return s.view.Close()
}

func (s *Namespace) IsClosed() bool {
	return s.view.IsClosed()
}

func newNamespace(repo *RepositoryImpl, name string, main bool) (*Namespace, error) {
	r, err := repo.View()
	if err != nil {
		return nil, err
	}
	container := &NamespaceContainer{
		repo:      r,
		namespace: name,
	}
	container.refs = accessio.NewRefCloser(container, true)
	container.ArtefactSetAccess = support.NewArtefactSetAccess(container)
	return container.view(main)
}

type NamespaceContainer struct {
	refs              accessio.ReferencableCloser
	repo              *Repository
	namespace         string
	ArtefactSetAccess *support.ArtefactSetAccess
}

var (
	_ support.ArtefactSetContainer = (*NamespaceContainer)(nil)
	_ cpi.NamespaceAccess          = (*Namespace)(nil)
)

func (a *NamespaceContainer) View(main ...bool) (support.ArtefactSetContainer, error) {
	ns, err := a.view(main...)
	if err != nil || ns == nil {
		return nil, err
	}
	return ns, err
}

func (a *NamespaceContainer) view(main ...bool) (*Namespace, error) {
	v, err := a.refs.View(main...)
	if err != nil {
		return nil, err
	}
	return &Namespace{view: v, NamespaceContainer: a}, nil
}

func (n *NamespaceContainer) GetNamespace() string {
	return n.namespace
}

func (n *NamespaceContainer) IsReadOnly() bool {
	return n.repo.IsReadOnly()
}

func (n *NamespaceContainer) IsClosed() bool {
	return n.repo.IsClosed()
}

func (n *NamespaceContainer) Close() error {
	return n.repo.Close()
}

func (n *NamespaceContainer) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return nil
}

func (n *NamespaceContainer) ListTags() ([]string, error) {
	return n.repo.getIndex().GetTags(n.namespace), nil // return digests as tags, also
}

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return n.repo.base.GetBlobData(digest)
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	n.repo.base.Lock()
	defer n.repo.base.Unlock()

	return n.repo.base.AddBlob(blob)
}

func (n *NamespaceContainer) GetArtefact(vers string) (cpi.ArtefactAccess, error) {
	meta := n.repo.getIndex().GetArtefactInfo(n.namespace, vers)
	if meta == nil {
		return nil, errors.ErrNotFound(cpi.KIND_OCIARTEFACT, vers, n.namespace)
	}
	return n.repo.base.GetArtefact(n, meta.Digest)
}

func (n *NamespaceContainer) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	n.repo.base.Lock()
	defer n.repo.base.Unlock()

	blob, err := n.repo.base.AddArtefactBlob(artefact)
	if err != nil {
		return nil, err
	}
	n.repo.getIndex().AddArtefactInfo(&index.ArtefactMeta{
		Repository: n.namespace,
		Tag:        "",
		Digest:     blob.Digest(),
	})
	return blob, n.AddTags(blob.Digest(), tags...)
}

func (n *NamespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	return n.repo.getIndex().AddTagsFor(n.namespace, digest, tags...)
}

////////////////////////////////////////////////////////////////////////////////

func (n *NamespaceContainer) GetRepository() cpi.Repository {
	return n.repo
}

func (n *NamespaceContainer) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if n.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return support.NewArtefact(n, art...)
}
