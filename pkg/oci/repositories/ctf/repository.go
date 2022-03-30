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
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/datacontext/vfsattr"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/artefactset"
	"github.com/gardener/ocm/pkg/oci/repositories/ctf/index"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

/*
   A common transport archive is just a folder with artefact archives.
   in tar format and an index.json file. The name of the archive
   is the digest of the artefact descriptor.

   The artefact archive is a filesystem structure with a file
   artefact-descriptor.json and a folder blobs containing
   the flat blob files with the name according to the blob digest.

   Digests used as filename will replace the ":" by a "."
*/

type Repository struct {
	base *artefactset.FileSystemBlobAccess
	spec *RepositorySpec
	ctx  cpi.Context
}

var _ cpi.Repository = &Repository{}

// New returns a new representation based repository
func New(ctx cpi.Context, spec *RepositorySpec, setup accessobj.Setup, closer accessobj.Closer, mode vfs.FileMode) (*Repository, error) {
	if spec.PathFileSystem == nil {
		spec.PathFileSystem = vfsattr.Get(ctx)
	}
	base, err := accessobj.NewAccessObject(accessObjectInfo, spec.AccessMode, spec.Options.Representation, setup, closer, mode)
	return _Wrap(ctx, spec, base, err)
}

func _Wrap(ctx cpi.Context, spec *RepositorySpec, obj *accessobj.AccessObject, err error) (*Repository, error) {
	if err != nil {
		return nil, err
	}
	r := &Repository{
		base: artefactset.NewFileSystemBlobAccess(obj),
		ctx:  ctx,
		spec: spec,
	}
	return r, nil
}

func (r *Repository) GetSpecification() cpi.RepositorySpec {
	return r.spec
}

func (r *Repository) NamespaceLister() cpi.NamespaceLister {
	return r
}

func (r *Repository) NumNamespaces(prefix string) (int, error) {
	return len(cpi.FilterByNamespacePrefix(prefix, r.getIndex().RepositoryList())), nil
}

func (r *Repository) GetNamespaces(prefix string, closure bool) ([]string, error) {
	return cpi.FilterChildren(closure, cpi.FilterByNamespacePrefix(prefix, r.getIndex().RepositoryList())), nil
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (r *Repository) IsReadOnly() bool {
	return r.base.IsReadOnly()
}

func (r *Repository) IsClosed() bool {
	return r.base.IsClosed()
}

func (r *Repository) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	return r.base.Write(path, mode, opts...)
}

func (r *Repository) Update() error {
	return r.base.Update()
}

func (r *Repository) Close() error {
	return r.base.Close()
}

func (a *Repository) getIndex() *index.RepositoryIndex {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*index.RepositoryIndex)
	}
	return a.base.GetState().GetState().(*index.RepositoryIndex)
}

////////////////////////////////////////////////////////////////////////////////
// cpi.Repository methods

func (r *Repository) ExistsArtefact(name string, tag string) (bool, error) {
	return r.getIndex().HasArtefact(name, tag), nil
}

func (r *Repository) LookupArtefact(name string, tag string) (cpi.ArtefactAccess, error) {
	a := r.getIndex().GetArtefactInfo(name, tag)
	if a == nil {
		return nil, cpi.ErrUnknownArtefact(name, tag)
	}

	return NewNamespace(r, name).GetArtefact(tag)
}

func (r *Repository) LookupNamespace(name string) (cpi.NamespaceAccess, error) {
	return NewNamespace(r, name), nil
}
