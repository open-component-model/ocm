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

package ocireg

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/errdefs"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/docker/resolve"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Namespace struct {
	access *NamespaceContainer
}

type NamespaceContainer struct {
	repo      *Repository
	namespace string
	resolver  resolve.Resolver
	lister    resolve.Lister
	fetcher   resolve.Fetcher
	pusher    resolve.Pusher
	blobs     *BlobContainers
}

var (
	_ cpi.ArtefactSetContainer = (*NamespaceContainer)(nil)
	_ cpi.NamespaceAccess      = (*Namespace)(nil)
)

func NewNamespace(repo *Repository, name string) (*Namespace, error) {
	ref := repo.getRef(name, "")
	resolver, err := repo.getResolver(name)
	if err != nil {
		return nil, err
	}
	fetcher, err := resolver.Fetcher(context.Background(), ref)
	if err != nil {
		return nil, err
	}
	pusher, err := resolver.Pusher(context.Background(), ref)
	if err != nil {
		return nil, err
	}
	lister, err := resolver.Lister(context.Background(), ref)
	if err != nil {
		return nil, err
	}
	n := &Namespace{
		access: &NamespaceContainer{
			repo:      repo,
			namespace: name,
			resolver:  resolver,
			lister:    lister,
			fetcher:   fetcher,
			pusher:    pusher,
			blobs:     NewBlobContainers(repo.ctx, fetcher, pusher),
		},
	}
	return n, nil
}

func (n *NamespaceContainer) Close() error {
	return n.blobs.Release()
}

func (n *NamespaceContainer) getPusher(vers string) (resolve.Pusher, error) {
	ref := n.repo.getRef(n.namespace, vers)
	resolver := n.resolver

	logrus.Infof("pusher for %s", ref)

	if ok, _ := artdesc.IsDigest(vers); !ok {
		var err error

		resolver, err = n.repo.getResolver(ref)

		if err != nil {
			return nil, fmt.Errorf("unable get resolver: %w", err)
		}
	}

	return resolver.Pusher(dummyContext, ref)
}

func (n *NamespaceContainer) push(vers string, blob cpi.BlobAccess) error {
	p, err := n.getPusher(vers)
	if err != nil {
		return fmt.Errorf("unable to get pusher: %w", err)
	}

	logrus.Infof("pushing %s", vers)

	return push(dummyContext, p, blob)
}

func (n *NamespaceContainer) GetNamepace() string {
	return n.namespace
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

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return n.blobs.Get("").GetBlobData(digest)
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	if _, _, err := n.blobs.Get("").AddBlob(blob); err != nil {
		return fmt.Errorf("unable to add blob: %w", err)
	}

	return nil
}

func (n *NamespaceContainer) ListTags() ([]string, error) {
	return n.lister.List(dummyContext)
}

func (n *NamespaceContainer) GetArtefact(vers string) (cpi.ArtefactAccess, error) {
	ref := n.repo.getRef(n.namespace, vers)
	logrus.Debugf("resolve %s\n", ref)
	_, desc, err := n.resolver.Resolve(context.Background(), ref)
	logrus.Debugf("done\n")
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, errors.ErrNotFound(cpi.KIND_OCIARTEFACT, ref, n.namespace)
		}
		return nil, err
	}
	_, acc, err := n.blobs.Get(desc.MediaType).GetBlobData(desc.Digest)
	if err != nil {
		return nil, err
	}
	return cpi.NewArtefactForBlob(n, accessio.BlobAccessForDataAccess(desc.Digest, desc.Size, desc.MediaType, acc))
}

func (n *NamespaceContainer) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	blob, err := artefact.Blob()
	if err != nil {
		return nil, err
	}
	if n.repo.info.Legacy {
		blob = artdesc.MapArtefactBlobMimeType(blob, true)
	}
	_, _, err = n.blobs.Get(blob.MimeType()).AddBlob(blob)
	if err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		for _, tag := range tags {
			err := n.push(tag, blob)
			if err != nil {
				return nil, err
			}
		}
	}

	return blob, err
}

func (n *NamespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	_, desc, err := n.resolver.Resolve(context.Background(), n.repo.getRef(n.namespace, digest.String()))
	if err != nil {
		return fmt.Errorf("unable to resolve: %w", err)
	}

	acc, err := NewDataAccess(n.fetcher, desc.Digest, desc.MediaType, false)
	if err != nil {
		return fmt.Errorf("error creating new data access: %w", err)
	}

	blob := accessio.BlobAccessForDataAccess(desc.Digest, desc.Size, desc.MediaType, acc)
	for _, tag := range tags {
		err := n.push(tag, blob)
		if err != nil {
			return fmt.Errorf("unable to push: %w", err)
		}
	}

	return nil
}

func (n *NamespaceContainer) NewArtefactProvider(state accessobj.State) (cpi.ArtefactProvider, error) {
	return cpi.NewNopCloserArtefactProvider(n), nil
}

////////////////////////////////////////////////////////////////////////////////

func (n *Namespace) Close() error {
	return n.access.Close()
}

func (n *Namespace) GetRepository() cpi.Repository {
	return n.access.repo
}

func (n *Namespace) GetNamespace() string {
	return n.access.GetNamepace()
}

func (n *Namespace) ListTags() ([]string, error) {
	return n.access.ListTags()
}

func (n *Namespace) NewArtefact(art ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if n.access.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return cpi.NewArtefact(n.access, art...)
}

func (n *Namespace) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return n.access.GetBlobData(digest)
}

func (n *Namespace) GetArtefact(vers string) (cpi.ArtefactAccess, error) {
	return n.access.GetArtefact(vers)
}

func (n *Namespace) AddArtefact(artefact cpi.Artefact, tags ...string) (accessio.BlobAccess, error) {
	return n.access.AddArtefact(artefact, tags...)
}

func (n *Namespace) AddTags(digest digest.Digest, tags ...string) error {
	return n.access.AddTags(digest, tags...)
}

func (n *Namespace) AddBlob(blob cpi.BlobAccess) error {
	return n.access.AddBlob(blob)
}
