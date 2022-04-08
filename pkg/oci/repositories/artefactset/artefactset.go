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
	"strings"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/oci/cpi"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
)

const MAINARTEFACT_ANNOTATION = "cloud.gardener.ocm/main"
const TAGS_ANNOTATION = "cloud.gardener.ocm/tags"
const TYPE_ANNOTATION = "cloud.gardener.ocm/type"

type ArtefactSet struct {
	base *FileSystemBlobAccess
	*cpi.ArtefactSetAccess
}

var _ cpi.ArtefactSetContainer = (*ArtefactSet)(nil)
var _ cpi.ArtefactSink = (*ArtefactSet)(nil)
var _ cpi.NamespaceAccess = (*ArtefactSet)(nil)

// New returns a new representation based element
func New(acc accessobj.AccessMode, fs vfs.FileSystem, setup accessobj.Setup, closer accessobj.Closer, mode vfs.FileMode) (*ArtefactSet, error) {
	return _Wrap(accessobj.NewAccessObject(accessObjectInfo, acc, fs, setup, closer, mode))
}

func _Wrap(obj *accessobj.AccessObject, err error) (*ArtefactSet, error) {
	if err != nil {
		return nil, err
	}
	s := &ArtefactSet{
		base: NewFileSystemBlobAccess(obj),
	}
	s.ArtefactSetAccess = cpi.NewArtefactSetAccess(s)
	return s, nil
}

func (a *ArtefactSet) GetNamespace() string {
	return ""
}

func (a *ArtefactSet) Annotate(name string, value string) {
	a.base.Lock()
	defer a.base.Unlock()

	d := a.GetIndex()
	if d.Annotations == nil {
		d.Annotations = map[string]string{}
	}
	d.Annotations[name] = value
}

////////////////////////////////////////////////////////////////////////////////
// sink

func (a *ArtefactSet) AddTags(digest digest.Digest, tags ...string) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()

	idx := a.GetIndex()
	for i, e := range idx.Manifests {
		if e.Digest == digest {
			if e.Annotations == nil {
				e.Annotations = map[string]string{}
				idx.Manifests[i].Annotations = e.Annotations
			}
			cur := e.Annotations[TAGS_ANNOTATION]
			if cur != "" {
				cur = strings.Join(append([]string{cur}, tags...), ",")
			} else {
				cur = strings.Join(tags, ",")
			}
			if cur != "" {
				e.Annotations[TAGS_ANNOTATION] = cur
			}
			return nil
		}
	}
	return errors.ErrUnknown(cpi.KIND_OCIARTEFACT, digest.String())
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *ArtefactSet) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *ArtefactSet) IsClosed() bool {
	return a.base.IsClosed()
}

func (a *ArtefactSet) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *ArtefactSet) Update() error {
	return a.base.Update()
}

func (a *ArtefactSet) Close() error {
	return a.base.Close()
}

// GetIndex returns the index of the included artefacts
// (image manifests and image indices)
// The manifst entries may describe dedicated tags
// to use for the dedicated artefact as annotation
// with the key TAGS_ANNOTATION.
func (a *ArtefactSet) GetIndex() *artdesc.Index {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.Index)
	}
	return a.base.GetState().GetState().(*artdesc.Index)
}

// GetMain returns the digest of the main artefact
// described by this artefact set.
// There might be more, if the main artefact is an index.
func (a *ArtefactSet) GetMain() digest.Digest {
	idx := a.GetIndex()
	if idx.Annotations == nil {
		return ""
	}
	return digest.Digest(idx.Annotations[MAINARTEFACT_ANNOTATION])
}

func (a *ArtefactSet) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return a.GetIndex().GetBlobDescriptor(digest)
}

func (a *ArtefactSet) GetBlobData(digest digest.Digest) (cpi.DataAccess, error) {
	return a.base.GetBlobData(digest)
}

func (a *ArtefactSet) AddBlob(blob cpi.BlobAccess) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return accessio.ErrReadOnly
	}
	if blob == nil {
		return nil
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.base.AddBlob(blob)
}

func (a *ArtefactSet) ListTags() ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Annotations != nil {
			if tags, ok := a.Annotations[TAGS_ANNOTATION]; ok {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *ArtefactSet) GetTags(digest digest.Digest) ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Digest == digest && a.Annotations != nil {
			if tags, ok := a.Annotations[TAGS_ANNOTATION]; ok {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *ArtefactSet) HasArtefact(ref string) (bool, error) {
	if a.IsClosed() {
		return false, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.hasArtefact(ref)
}

func (a *ArtefactSet) GetArtefact(ref string) (cpi.ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.getArtefact(ref)
}

func (a *ArtefactSet) matcher(ref string) func(d *artdesc.Descriptor) bool {
	if ok, digest := artdesc.IsDigest(ref); ok {
		return func(desc *artdesc.Descriptor) bool {
			return desc.Digest == digest
		}
	}
	return func(d *artdesc.Descriptor) bool {
		if d.Annotations == nil {
			return false
		}
		for _, tag := range strings.Split(d.Annotations[TAGS_ANNOTATION], ",") {
			if tag == ref {
				return true
			}
		}
		return false
	}
}

func (a *ArtefactSet) hasArtefact(ref string) (bool, error) {
	idx := a.GetIndex()
	match := a.matcher(ref)
	for _, e := range idx.Manifests {
		if match(&e) {
			return true, nil
		}
	}
	return false, nil
}

func (a *ArtefactSet) getArtefact(ref string) (cpi.ArtefactAccess, error) {
	idx := a.GetIndex()
	match := a.matcher(ref)
	for _, e := range idx.Manifests {
		if match(&e) {
			return a.base.GetArtefact(a, e.Digest)
		}
	}
	return nil, errors.ErrUnknown(cpi.KIND_OCIARTEFACT, ref)
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

func (a *ArtefactSet) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	blob, err := a.AddPlatformArtefact(artefact, nil)
	if err != nil {
		return nil, err
	}
	return blob, a.AddTags(blob.Digest(), tags...)
}

func (a *ArtefactSet) AddPlatformArtefact(artefact cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
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

func (a *ArtefactSet) NewArtefact(artefact ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return cpi.NewArtefact(a, artefact...)
}

func (a *ArtefactSet) NewArtefactProvider(state accessobj.State) (cpi.ArtefactProvider, error) {
	return cpi.NewNopCloserArtefactProvider(a), nil
}
