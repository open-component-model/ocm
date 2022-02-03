// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ctf

import (
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/artefactset"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/index"
	"github.com/opencontainers/go-digest"
)

type Namespace struct {
	access *NamespaceContainer
}

type NamespaceContainer struct {
	repo              *Repository
	namespace         string
	ArtefactSetAccess *artefactset.ArtefactSetAccess
}

var _ artefactset.ArtefactSetContainer = (*NamespaceContainer)(nil)
var _ cpi.NamespaceAccess = (*Namespace)(nil)

func NewNamespace(repo *Repository, name string) *Namespace {
	n := &Namespace{
		access: &NamespaceContainer{
			repo:      repo,
			namespace: name,
		},
	}
	n.access.ArtefactSetAccess = artefactset.NewArtefactSetAccess(n.access)
	return n
}

func (n *NamespaceContainer) IsReadOnly() bool {
	return n.repo.IsReadOnly()
}

func (n *NamespaceContainer) IsClosed() bool {
	return n.repo.IsClosed()
}

func (n *NamespaceContainer) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return nil
}

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	return n.repo.base.GetBlobData(digest)
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	n.repo.base.Lock()
	defer n.repo.base.Unlock()

	return n.repo.base.AddBlob(blob)
}

func (n *NamespaceContainer) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	for _, a := range n.repo.getIndex().GetArtefacts(digest) {
		if a.Repository == n.namespace {
			return n.repo.base.GetArtefact(n, digest)
		}
	}
	return nil, errors.ErrNotFound(cpi.KIND_OCIARTEFACT, digest.String(), n.namespace)
}

func (n *NamespaceContainer) AddArtefact(artefact cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	n.repo.base.Lock()
	defer n.repo.base.Unlock()

	blob, err := n.repo.base.AddArtefactBlob(artefact)
	if err != nil {
		return nil, err
	}
	n.repo.getIndex().AddArtefact(&index.ArtefactMeta{
		Repository: n.namespace,
		Tag:        "",
		Digest:     blob.Digest(),
	})
	return blob, nil
}

func (n *Namespace) GetRepository() cpi.Repository {
	return n.access.repo
}

func (n *Namespace) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	return artefactset.NewArtefact(n.access, art...), nil
}

func (n *Namespace) GetArtefactByTag(tag string) (cpi.ArtefactAccess, error) {
	meta := n.access.repo.getIndex().GetArtefact(n.access.namespace, tag)
	if meta != nil {
		return n.access.repo.base.GetArtefact(n.access, meta.Digest)
	}
	return nil, errors.ErrNotFound(cpi.KIND_OCIARTEFACT, tag, n.access.namespace)
}

func (n *Namespace) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	return n.access.GetArtefact(digest)
}

func (n *Namespace) AddArtefact(artefact cpi.Artefact) (access accessio.BlobAccess, err error) {
	return n.access.AddArtefact(artefact, nil)
}
