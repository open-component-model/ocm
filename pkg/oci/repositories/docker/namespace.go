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

package docker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type Namespace struct {
	access *NamespaceContainer
}

func (n *Namespace) Close() error {
	return nil
}

type NamespaceContainer struct {
	repo      *Repository
	namespace string
}

var _ cpi.ArtefactSetContainer = (*NamespaceContainer)(nil)
var _ cpi.NamespaceAccess = (*Namespace)(nil)

func NewNamespace(repo *Repository, name string) (*Namespace, error) {
	n := &Namespace{
		access: &NamespaceContainer{
			repo:      repo,
			namespace: name,
		},
	}
	return n, nil
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

func (n *NamespaceContainer) ListTags() ([]string, error) {
	opts := dockertypes.ImageListOptions{}
	list, err := n.repo.client.ImageList(dummyContext, opts)
	if err != nil {
		return nil, err
	}
	var result []string
	if n.namespace == "" {
		for _, e := range list {
			result = append(result, e.ID)
		}
	} else {
		prefix := n.namespace + ":"
		for _, e := range list {
			for _, t := range e.RepoTags {
				if strings.HasPrefix(t, prefix) {
					result = append(result, t[len(prefix):])
				}
			}
		}
	}
	return result, nil
}

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	return nil, errors.ErrNotImplemented()
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	return accessio.ErrReadOnly
}

func (n *NamespaceContainer) GetArtefact(vers string) (cpi.ArtefactAccess, error) {
	ref, err := ParseRef(n.namespace, vers)
	if err != nil {
		return nil, err
	}
	src, err := ref.NewImageSource(dummyContext, nil)
	if err != nil {
		return nil, err
	}

	data, mime, err := src.GetManifest(dummyContext, nil)
	if err != nil {
		src.Close()
		return nil, err
	}

	fmt.Printf("mime: %s\n", mime)
	fmt.Printf("manifest:\n %s\n*********\n", string(data))

	opts := types.ManifestUpdateOptions{
		ManifestMIMEType: artdesc.MediaTypeImageManifest,
	}
	un := image.UnparsedInstance(src, nil)
	img, err := image.FromUnparsedImage(dummyContext, nil, un)
	if err != nil {
		src.Close()
		return nil, err
	}

	img, err = img.UpdatedImage(dummyContext, opts)
	if err != nil {
		src.Close()
		return nil, err
	}

	data, mime, err = img.Manifest(dummyContext)
	if err != nil {
		src.Close()
		return nil, err
	}

	p := &daemonArtefactProvider{
		namespace: n,
		src:       src,
		img:       img,
	}
	return cpi.NewArtefactForProviderBlob(n, p, accessio.BlobAccessForData(mime, data))
}

func (n *NamespaceContainer) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	return nil, accessio.ErrReadOnly
}

func (n *NamespaceContainer) AddTags(digest digest.Digest, tags ...string) error {
	return accessio.ErrReadOnly
}

func (n *NamespaceContainer) NewArtefactProvider(state accessobj.State) (cpi.ArtefactProvider, error) {
	return nil, nil
}

type daemonArtefactProvider struct {
	lock      sync.Mutex
	namespace *NamespaceContainer
	src       types.ImageSource
	img       types.Image
}

var _ cpi.ArtefactProvider = (*daemonArtefactProvider)(nil)

func (d *daemonArtefactProvider) IsClosed() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.src == nil
}

func (d *daemonArtefactProvider) IsReadOnly() bool {
	return true
}

func (d *daemonArtefactProvider) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return nil
}

func (d *daemonArtefactProvider) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.src != nil {
		err := d.src.Close()
		d.src = nil
		return err
	}
	return nil
}

func (d *daemonArtefactProvider) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	info := d.img.ConfigInfo()
	if info.Digest == digest {
		data, err := d.img.ConfigBlob(dummyContext)
		if err != nil {
			return nil, err
		}
		return accessio.DataAccessForBytes(data), nil
	}
	info.Digest = ""
	for _, l := range d.img.LayerInfos() {
		if l.Digest == digest {
			info = l
			return NewDataAccess(d.src, info, false)
		}
	}
	return nil, cpi.ErrBlobNotFound(digest)
}

func (d *daemonArtefactProvider) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	return nil, errors.ErrInvalid()
}

func (d *daemonArtefactProvider) AddBlob(access cpi.BlobAccess) error {
	return accessio.ErrReadOnly
}

func (d *daemonArtefactProvider) AddArtefact(art cpi.Artefact) (access accessio.BlobAccess, err error) {
	return nil, accessio.ErrReadOnly
}

////////////////////////////////////////////////////////////////////////////////

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
	return nil, errors.ErrNotImplemented()
}

func (n *Namespace) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
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
