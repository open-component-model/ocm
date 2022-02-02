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
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/opencontainers/go-digest"
)

type Artefact struct {
	lock sync.RWMutex
	set  *ArtefactSet
	*artdesc.Artefact
	*BlobContainer
}

func NewArtefact(set *ArtefactSet, artefact *artdesc.Artefact) *Artefact {
	if artefact == nil {
		artefact = artdesc.New()
	}
	a := &Artefact{
		set:      set,
		Artefact: artefact,
	}
	a.BlobContainer = NewBlobContainer(set, a)
	return a
}

func (a *Artefact) GetDescriptor() *artdesc.Artefact {
	return a.Artefact
}

func (a *Artefact) IsClosed() bool {
	return a.set.IsClosed()
}

func (a *Artefact) IsReadOnly() bool {
	return a.set.IsReadOnly()
}

func (i *Artefact) GetBlob(digest digest.Digest) (cpi.BlobAccess, error) {
	d := i.GetBlobDescriptor(digest)
	if d != nil {
		data, err := i.set.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForDataAccess(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, cpi.ErrBlobNotFound(digest)
}

func (a *Artefact) AddManifest(manifest *artdesc.Manifest, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	idx := a.GetDescriptor().Index()
	if idx == nil {
		idx = artdesc.NewIndex()
		err := a.GetDescriptor().SetIndex(idx)
		if err != nil {
			return nil, err
		}
	}
	blob, err := manifest.ToBlobAccess()
	if err != nil {
		return nil, err
	}

	err = a.AddBlob(blob)
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
	a.lock.Lock()
	defer a.lock.Unlock()
	m := a.GetDescriptor().Manifest()
	if m == nil {
		m = artdesc.NewManifest()
		err := a.GetDescriptor().SetManifest(m)
		if err != nil {
			return -1, err
		}
	}

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

	err := a.set.AddBlob(blob)
	if err != nil {
		return -1, err
	}

	m.Layers = append(m.Layers, *d)
	return len(m.Layers) - 1, nil
}
