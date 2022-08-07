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

package comparch

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localfsblob"
	ocmhdlr "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////

// ComponentArchive is the go representation for a component artefact
type ComponentArchive struct {
	base *accessobj.FileSystemBlobAccess
	comp *ComponentAccess
	*support.ComponentVersionAccess
}

var _ support.ComponentVersionContainer = (*ComponentArchive)(nil)

// New returns a new representation based element
func New(ctx cpi.Context, acc accessobj.AccessMode, fs vfs.FileSystem, setup accessobj.Setup, closer accessobj.Closer, mode vfs.FileMode) (*ComponentArchive, error) {
	obj, err := accessobj.NewAccessObject(accessObjectInfo, acc, fs, setup, closer, mode)
	return _Wrap(ctx, obj, NewRepositorySpec(acc, ""), err)
}

func _Wrap(ctx cpi.Context, obj *accessobj.AccessObject, spec *RepositorySpec, err error) (*ComponentArchive, error) {
	if err != nil {
		return nil, err
	}
	s := &ComponentArchive{
		base: accessobj.NewFileSystemBlobAccess(obj),
	}
	s.comp = &ComponentAccess{&Repository{
		ctx:  ctx,
		spec: spec,
		arch: s,
	}}
	s.ComponentVersionAccess, err = support.NewComponentVersionAccess(s, false)
	if err != nil {
		return nil, err
	}
	return s, nil
}

////////////////////////////////////////////////////////////////////////////////

var _ cpi.ComponentVersionAccess = &ComponentArchive{}

func (c *ComponentArchive) GetContext() cpi.Context {
	return c.comp.repo.GetContext()
}

func (c *ComponentArchive) AsRepository() cpi.Repository {
	return c.comp.repo
}

func (c *ComponentArchive) Repository() cpi.Repository {
	return c.comp.repo
}

func (c *ComponentArchive) ComponentAccess() cpi.ComponentAccess {
	return c.comp
}

func (c *ComponentArchive) Update() error {
	return c.base.Update()
}

func (c *ComponentArchive) Close() error {
	return c.base.Close()
}

func (c *ComponentArchive) SetName(n string) {
	c.GetDescriptor().Name = n
}

func (c *ComponentArchive) SetVersion(v string) {
	c.GetDescriptor().Version = v
}

func (c *ComponentArchive) AccessMethod(a cpi.AccessSpec) (cpi.AccessMethod, error) {
	if a.GetKind() == localblob.Type || a.GetKind() == localfsblob.Type {
		a, err := c.GetContext().AccessSpecForSpec(a)
		if err != nil {
			return nil, err
		}
		return newLocalFilesystemBlobAccessMethod(a.(*localblob.AccessSpec), c)
	}
	return nil, errors.ErrNotSupported(errors.KIND_ACCESSMETHOD, a.GetType(), "component archive")
}

func (c *ComponentArchive) GetBlobData(name string) (cpi.DataAccess, error) {
	return c.base.GetBlobDataByName(name)
}

func (c *ComponentArchive) GetStorageContext(cv cpi.ComponentVersionAccess) cpi.StorageContext {
	return ocmhdlr.New(c.AsRepository(), cv, c.base, Type)
}

func (c *ComponentArchive) AddBlobFor(storagectx cpi.StorageContext, blob cpi.BlobAccess, refName string, global cpi.AccessSpec) (cpi.AccessSpec, error) {
	if blob == nil {
		return nil, errors.New("a resource has to be defined")
	}
	err := c.base.AddBlob(blob)
	if err != nil {
		return nil, err
	}
	return localblob.New(common.DigestToFileName(blob.Digest()), refName, blob.MimeType(), global), nil
}

func (c *ComponentArchive) GetDescriptor() *compdesc.ComponentDescriptor {
	if c.base.IsReadOnly() {
		return c.base.GetState().GetOriginalState().(*compdesc.ComponentDescriptor)
	}
	return c.base.GetState().GetState().(*compdesc.ComponentDescriptor)
}
