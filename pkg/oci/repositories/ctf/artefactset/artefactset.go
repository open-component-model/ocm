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
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

const TAGS_ANNOTATION = "ocm.gardener.cloud/tags"
const TYPE_ANNOTATION = "ocm.gardener.cloud/type"

type ArtefactSet struct {
	base *FileSystemBlobAccess
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
		base: NewFileSystemBlobAccess(obj),
	}
	s.ArtefactSetAccess = NewArtefactSetAccess(s)
	return s, nil
}

////////////////////////////////////////////////////////////////////////////////
// forward

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

func (a *ArtefactSet) GetIndex() *artdesc.Index {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.Index)
	}
	return a.base.GetState().GetState().(*artdesc.Index)
}

func (a *ArtefactSet) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return a.GetIndex().GetBlobDescriptor(digest)
}

func (a *ArtefactSet) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	return a.base.GetBlobData(digest)
}

func (a *ArtefactSet) AddBlob(blob cpi.BlobAccess) error {
	a.base.Lock()
	defer a.base.Unlock()
	return a.base.AddBlob(blob)
}

func (i *ArtefactSet) GetArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	if i.IsClosed() {
		return nil, accessio.ErrClosed
	}
	i.base.Lock()
	defer i.base.Unlock()
	return i.getArtefact(digest)
}

func (i *ArtefactSet) getArtefact(digest digest.Digest) (cpi.ArtefactAccess, error) {
	idx := i.GetIndex()
	for _, e := range idx.Manifests {
		if e.Digest == digest {
			return i.base.GetArtefact(i, e.Digest)
		}
	}
	return nil, errors.ErrUnknown(cpi.KIND_OCIARTEFACT, digest.String())
}

func (a *ArtefactSet) AnnotateArtefact(digest digest.Digest, name, value string) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	a.base.Lock()
	defer a.base.Unlock()
	idx := a.GetIndex()
	for i, e := range idx.Manifests {
		if e.Digest == digest {
			annos := e.Annotations
			if annos == nil {
				annos = map[string]string{}
				idx.Manifests[i].Annotations = annos
			}
			annos[name] = value
			return nil
		}
	}
	return errors.ErrUnknown(cpi.KIND_OCIARTEFACT, digest.String())
}

func (a *ArtefactSet) AddArtefact(artefact cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	a.base.Lock()
	defer a.base.Unlock()
	idx := a.GetIndex()
	blob, err := a.base.AddArtefactBlob(artefact)
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
