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

package artefact

import (
	"compress/gzip"
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/core"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

type Artefact struct {
	base *accessobj.AccessObject

	lock      sync.RWMutex
	blobinfos map[digest.Digest]*cpi.Descriptor
}

var _ cpi.ArtefactAccess = (*struct {
	*Artefact
	cpi.RepositorySource
})(nil)

// New returns a new representation based element
func New(acc accessobj.AccessMode, fs vfs.FileSystem, closer accessobj.Closer, mode vfs.FileMode) (*Artefact, error) {
	return _Wrap(accessobj.NewAccessObject(accessObjectInfo, acc, fs, closer, mode))
}

func _Wrap(obj *accessobj.AccessObject, err error) (*Artefact, error) {
	if err != nil {
		return nil, err
	}
	return &Artefact{
		base:      obj,
		blobinfos: map[digest.Digest]*cpi.Descriptor{},
	}, nil
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *Artefact) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *Artefact) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *Artefact) Write(path string, mode vfs.FileMode, opts ...accessobj.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *Artefact) Update() error {
	return a.base.Update()
}

func (a *Artefact) Close() error {
	return a.base.Close()
}

////////////////////////////////////////////////////////////////////////////////
// Object functionality

////////////////////////////////////////////////////////////////////////////////
// methods for Access

func (a *Artefact) GetDescriptor() *artdesc.ArtefactDescriptor {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.ArtefactDescriptor)
	}
	return a.base.GetState().GetState().(*artdesc.ArtefactDescriptor)
}

func (a *Artefact) GetIndex(digest digest.Digest) (core.IndexAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	return NewBlobContainer(a, a).GetIndex(digest)
}

func (a *Artefact) GetManifest(digest digest.Digest) (core.ManifestAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	return NewBlobContainer(a, a).GetManifest(digest)
}

func (a *Artefact) GetBlob(digest digest.Digest) (core.BlobAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	d := a.GetBlobDescriptor(digest)
	if d != nil {
		data, err := a.GetBlobData(digest)
		if err != nil {
			return nil, err
		}
		return accessio.BlobAccessForFile(d.Digest, d.Size, d.MediaType, data), nil
	}
	return nil, errors.ErrNotFound("blob", string(digest))
}

func (a *Artefact) GetBlobData(digest digest.Digest) (core.DataAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	path := DigestPath(digest)
	if ok, err := vfs.FileExists(a.base.GetFileSystem(), path); ok {
		return accessio.DataAccessForFile(a.base.GetFileSystem(), path), nil
	} else {
		if err != nil {
			return nil, err
		}
		return nil, cpi.ErrBlobNotFound(digest)
	}
}

func (a *Artefact) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	a.lock.RLock()
	defer a.lock.RUnlock()

	d := a.blobinfos[digest]
	if d == nil {
		d = a.GetDescriptor().GetBlobDescriptor(digest)
	}
	return d
}

////////////////////////////////////////////////////////////////////////////////
// methods for Composer

var _ cpi.ArtefactComposer = (*struct {
	*Artefact
	cpi.RepositorySource
})(nil) // magic

func (a *Artefact) AddManifest(manifest *artdesc.ArtefactDescriptor, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
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

	err = a.addBlob(blob)
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

func (a *Artefact) AddLayer(blob core.BlobAccess, d *artdesc.Descriptor) (int, error) {
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

	err := a.addBlob(blob)
	if err != nil {
		return -1, err
	}

	m.Layers = append(m.Layers, *d)
	return len(m.Layers) - 1, nil
}

func (a *Artefact) AddBlob(blob core.BlobAccess) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	a.lock.Lock()
	defer a.lock.Unlock()

	return a.addBlob(blob)
}

func (a *Artefact) addBlob(blob core.BlobAccess) error {
	path := DigestPath(blob.Digest())
	if ok, err := vfs.FileExists(a.base.GetFileSystem(), path); ok {
		return nil
	} else {
		if err != nil {
			return err
		}
	}
	data, err := blob.Get()
	if err != nil {
		return err
	}
	err = vfs.WriteFile(a.base.GetFileSystem(), path, data, a.base.GetMode()&0666)
	if err != nil {
		return err
	}
	a.blobinfos[blob.Digest()] = &artdesc.Descriptor{
		MediaType: blob.MimeType(),
		Digest:    blob.Digest(),
		Size:      blob.Size(),
	}
	return nil
}
