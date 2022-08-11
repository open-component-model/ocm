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
	"strings"
	"sync"

	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
)

type NamespaceContainer struct {
	lock      sync.RWMutex
	repo      *Repository
	namespace string
	cache     accessio.BlobCache
}

var _ cpi.ArtefactSetContainer = (*NamespaceContainer)(nil)
var _ cpi.NamespaceAccess = (*Namespace)(nil)

func NewNamespace(repo *Repository, name string) (*Namespace, error) {
	cache, err := accessio.NewCascadedBlobCache(nil)
	if err != nil {
		return nil, err
	}
	n := &Namespace{
		access: &NamespaceContainer{
			repo:      repo,
			namespace: name,
			cache:     cache,
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
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.cache == nil
}

func (n *NamespaceContainer) Close() error {
	n.lock.Lock()
	defer n.lock.Unlock()
	if n.cache != nil {
		err := n.cache.Unref()
		n.cache = nil
		return err
	}
	return nil
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
			// ID is always the config digest
			// filter images without a repo tag for empty namespace
			if len(e.RepoTags) == 0 {
				d, err := digest.Parse(e.ID)
				if err == nil {
					result = append(result, d.String()[:12])
				}
			}
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

func (n *NamespaceContainer) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return n.cache.GetBlobData(digest)
}

func (n *NamespaceContainer) AddBlob(blob cpi.BlobAccess) error {
	_, _, err := n.cache.AddBlob(blob)
	return err
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

	/*
		data, mime, err := src.GetManifest(dummyContext, nil)
		if err != nil {
			src.Close()
			return nil, err
		}

		//fmt.Printf("mime: %s\n", mime)
		//fmt.Printf("manifest:\n %s\n*********\n", string(data))
	*/

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

	data, mime, err := img.Manifest(dummyContext)
	if err != nil {
		src.Close()
		return nil, err
	}

	cache, err := accessio.NewCascadedBlobCacheForSource(n.cache, newDockerSource(img, src))
	if err != nil {
		return nil, err
	}
	p := &daemonArtefactProvider{
		namespace: n,
		cache:     cache,
	}
	return cpi.NewArtefactForProviderBlob(n, p, accessio.BlobAccessForData(mime, data))
}

func (n *NamespaceContainer) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	tag := "latest"
	if len(tags) > 0 {
		tag = tags[0]
	}
	ref, err := ParseRef(n.namespace, tag)
	if err != nil {
		return nil, err
	}
	dst, err := ref.NewImageDestination(dummyContext, nil)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	blob, err := Convert(artefact, n.cache, dst)
	if err != nil {
		return nil, err
	}
	err = dst.Commit(dummyContext, nil)
	if err != nil {
		return nil, err
	}

	return blob, nil
}

func (n *NamespaceContainer) AddTags(digest digest.Digest, tags ...string) error {

	if ok, _ := artdesc.IsDigest(digest.String()); ok {
		return errors.ErrNotSupported("image access by digest")
	}
	src := n.namespace + ":" + digest.String()
	if pattern.MatchString(digest.String()) {
		// this definately no digest, but the library expects it this way
		src = digest.String()
	}
	for _, tag := range tags {
		err := n.repo.client.ImageTag(dummyContext, src, n.namespace+":"+tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *NamespaceContainer) NewArtefactProvider(state accessobj.State) (cpi.ArtefactProvider, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////////////////////////

type Namespace struct {
	access *NamespaceContainer
}

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
	var m *artdesc.Artefact
	if len(art) == 0 {
		m = artdesc.NewManifestArtefact()
	} else {
		if !art[0].IsManifest() {
			err := m.SetManifest(artdesc.NewManifest())
			if err != nil {
				return nil, err
			}
		}
		m = art[0]
	}
	return cpi.NewArtefact(n.access, m)
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
