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

package artefactset

import (
	"compress/gzip"
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

var ErrNoIndex = errors.New("manifest does not support access to subsequent artefacts")

type Artefact struct {
	lock     sync.RWMutex
	access   ArtefactSetContainer
	artefact *artdesc.Artefact
	handler  *BlobHandler
}

func NewArtefact(access ArtefactSetContainer, artefact *artdesc.Artefact) *Artefact {
	if artefact == nil {
		artefact = artdesc.New()
	}
	a := &Artefact{
		access:   access,
		artefact: artefact,
	}
	a.handler = NewBlobHandler(access, a)
	return a
}

func (a *Artefact) IsClosed() bool {
	return a.access.IsClosed()
}

func (a *Artefact) IsReadOnly() bool {
	return a.access.IsReadOnly()
}

////////////////////////////////////////////////////////////////////////////////
// from artdesc.Artefact

func (a *Artefact) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	d := a.artefact.GetBlobDescriptor(digest)
	if d != nil {
		return d
	}
	return a.access.GetBlobDescriptor(digest)
}

func (a *Artefact) IsIndex() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.artefact.IsIndex()
}

func (a *Artefact) IsManifest() bool {
	a.lock.RLock()
	defer a.lock.RUnlock()
	return a.artefact.IsManifest()
}

func (a *Artefact) Index() (*artdesc.Index, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	idx := a.artefact.Index()
	if idx == nil {
		idx = artdesc.NewIndex()
		if err := a.artefact.SetIndex(idx); err != nil {
			return nil, err
		}
	}
	return idx, nil
}

func (a *Artefact) Manifest() (*artdesc.Manifest, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	m := a.artefact.Manifest()
	if m == nil {
		m = artdesc.NewManifest()
		if err := a.artefact.SetManifest(m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (a *Artefact) ToBlobAccess() (cpi.BlobAccess, error) {
	return a.artefact.ToBlobAccess()
}

////////////////////////////////////////////////////////////////////////////////
// from BlobHandler

func (i *Artefact) GetArtefact(digest digest.Digest) (*Artefact, error) {
	if !i.IsIndex() {
		return nil, ErrNoIndex
	}
	return i.handler.GetArtefact(digest)
}

func (i *Artefact) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	return i.handler.GetBlob(digest)
}

func (i *Artefact) GetManifest(digest digest.Digest) (cpi.ManifestAccess, error) {
	if !i.IsIndex() {
		return nil, ErrNoIndex
	}
	return i.handler.GetManifest(digest)
}

func (i *Artefact) GetIndex(digest digest.Digest) (cpi.IndexAccess, error) {
	if !i.IsIndex() {
		return nil, ErrNoIndex
	}
	return i.handler.GetIndex(digest)
}

////////////////////////////////////////////////////////////////////////////////

func (a *Artefact) AddManifest(manifest *artdesc.Manifest, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	idx, err := a.Index()
	if err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	blob, err := manifest.ToBlobAccess()
	if err != nil {
		return nil, err
	}

	err = a.handler.AddBlob(blob)
	if err != nil {
		return nil, err
	}

	idx.Manifests = append(idx.Manifests, cpi.Descriptor{
		MediaType:   blob.MimeType(),
		Digest:      blob.Digest(),
		Size:        blob.Size(),
		URLs:        nil,
		Annotations: nil,
		Platform:    platform,
	})
	return blob, nil
}

func (a *Artefact) AddLayer(blob cpi.BlobAccess, d *artdesc.Descriptor) (int, error) {
	if a.IsClosed() {
		return -1, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return -1, accessio.ErrReadOnly
	}
	m, err := a.Manifest()
	if err != nil {
		return -1, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	if d == nil {
		d = &artdesc.Descriptor{}
	}
	d.Digest = blob.Digest()
	d.Size = blob.Size()
	if d.MediaType == "" {
		d.MediaType = blob.MimeType()
		if d.MediaType == "" {
			d.MediaType = artdesc.MediaTypeImageLayer
			r, err := blob.Reader()
			if err != nil {
				return -1, err
			}
			defer r.Close()
			zr, err := gzip.NewReader(r)
			if err == nil {
				err = zr.Close()
				if err == nil {
					d.MediaType = artdesc.MediaTypeImageLayerGzip
				}
			}
		}
	}

	err = a.access.AddBlob(blob)
	if err != nil {
		return -1, err
	}

	m.Layers = append(m.Layers, *d)
	return len(m.Layers) - 1, nil
}
