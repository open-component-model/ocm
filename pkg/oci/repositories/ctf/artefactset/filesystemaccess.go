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
	"sync"

	"github.com/gardener/ocm/pkg/common"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

type FileSystemBlobAccess struct {
	sync.RWMutex
	base *accessobj.AccessObject
}

func NewFileSystemBlobAccess(access *accessobj.AccessObject) *FileSystemBlobAccess {
	return &FileSystemBlobAccess{
		base: access,
	}
}

func (a *FileSystemBlobAccess) Access() *accessobj.AccessObject {
	return a.base
}

func (a *FileSystemBlobAccess) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *FileSystemBlobAccess) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *FileSystemBlobAccess) Write(path string, mode vfs.FileMode, opts ...accessobj.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *FileSystemBlobAccess) Update() error {
	return a.base.Update()
}

func (a *FileSystemBlobAccess) Close() error {
	return a.base.Close()
}

func (a *FileSystemBlobAccess) GetState() accessobj.State {
	return a.base.GetState()
}

// DigestPath returns the path to the blob for a given name.
func (a *FileSystemBlobAccess) DigestPath(digest digest.Digest) string {
	return a.base.GetInfo().SubPath(common.DigestToFileName(digest))
}

func (a *FileSystemBlobAccess) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
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

func (a *FileSystemBlobAccess) AddBlob(blob cpi.BlobAccess) error {
	if a.base.IsClosed() {
		return accessio.ErrClosed
	}
	if a.base.IsReadOnly() {
		return accessio.ErrReadOnly
	}

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
	return vfs.WriteFile(a.base.GetFileSystem(), path, data, a.base.GetMode()&0666)
}

func (i *FileSystemBlobAccess) getArtefact(blob cpi.BlobAccess) (*artdesc.Artefact, error) {
	data, err := blob.Get()
	if err != nil {
		return nil, err
	}
	return artdesc.Decode(data)
}

func (i *FileSystemBlobAccess) GetArtefact(access ArtefactSetContainer, digest digest.Digest) (cpi.ArtefactAccess, error) {
	data, err := i.GetBlobData(digest)
	if err != nil {
		return nil, err
	}

	d, err := i.getArtefact(accessio.BlobAccessForDataAccess("", -1, "", data))
	if err != nil {
		return nil, err
	}
	return NewArtefact(access, d), nil
}

func (i *FileSystemBlobAccess) AddArtefactBlob(artefact cpi.Artefact) (cpi.BlobAccess, error) {
	blob, err := artefact.Artefact().ToBlobAccess()
	if err != nil {
		return nil, err
	}

	err = i.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return blob, nil
}
