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
	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

type ArtefactSet struct {
	base *accessobj.AccessObject
	*ArtefactSetAccess
}

var _ ArtefactSetContainer = (*ArtefactSet)(nil)

// New returns a new representation based element
func New(acc accessobj.AccessMode, fs vfs.FileSystem, closer accessobj.Closer, mode vfs.FileMode) (*ArtefactSet, error) {
	return _Wrap(accessobj.NewAccessObject(accessObjectInfo, acc, fs, closer, mode))
}

func _Wrap(obj *accessobj.AccessObject, err error) (*ArtefactSet, error) {
	if err != nil {
		return nil, err
	}
	s := &ArtefactSet{
		base: obj,
	}
	s.ArtefactSetAccess = NewArtefactSetAccess(s)
	return s, nil
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *ArtefactSet) GetFileSystem() vfs.FileSystem {
	return a.base.GetFileSystem()
}

func (a *ArtefactSet) GetInfo() *accessobj.AccessObjectInfo {
	return a.base.GetInfo()
}

func (a *ArtefactSet) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *ArtefactSet) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *ArtefactSet) Write(path string, mode vfs.FileMode, opts ...accessobj.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *ArtefactSet) Update() error {
	return a.base.Update()
}

func (a *ArtefactSet) Close() error {
	return a.base.Close()
}

// DigestPath returns the path to the blob for a given name.
func (a *ArtefactSet) DigestPath(digest digest.Digest) string {
	return a.base.GetInfo().SubPath(common.DigestToFileName(digest))
}

func (a *ArtefactSet) GetDescriptor() *artdesc.Index {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.Index)
	}
	return a.base.GetState().GetState().(*artdesc.Index)
}

func (a *ArtefactSet) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return a.GetDescriptor().GetBlobDescriptor(digest)
}

func (a *ArtefactSet) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	path := a.DigestPath(digest)
	if ok, err := vfs.FileExists(a.base.GetFileSystem(), path); ok {
		return accessio.DataAccessForFile(a.base.GetFileSystem(), path), nil
	} else {
		if err != nil {
			return nil, err
		}
		return nil, cpi.ErrBlobNotFound(digest)
	}
}

func (a *ArtefactSet) AddBlob(blob cpi.BlobAccess) error {
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

func (a *ArtefactSet) addBlob(blob cpi.BlobAccess) error {
	path := a.DigestPath(blob.Digest())
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

func (i *ArtefactSet) getArtefact(blob cpi.BlobAccess) (*artdesc.Artefact, error) {
	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	return artdesc.Decode(data)
}

func (i *ArtefactSet) GetArtefact(digest digest.Digest) (*Artefact, error) {
	blob, err := i.GetBlob(digest)
	if err != nil {
		return nil, err
	}

	d, err := i.getArtefact(blob)
	if err != nil {
		return nil, err
	}
	return NewArtefact(i.set, d), nil
}

func (a *ArtefactSet) AddArtefact(artefact *Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	idx := a.GetDescriptor()
	blob, err := artefact.ToBlobAccess()
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
