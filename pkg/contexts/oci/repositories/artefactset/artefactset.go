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

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/oci/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
)

const (
	MAINARTEFACT_ANNOTATION = "software.ocm/main"
	TAGS_ANNOTATION         = "software.ocm/tags"
	TYPE_ANNOTATION         = "software.ocm/type"

	LEGACY_MAINARTEFACT_ANNOTATION = "cloud.gardener.ocm/main"
	LEGACY_TAGS_ANNOTATION         = "cloud.gardener.ocm/tags"
	LEGACY_TYPE_ANNOTATION         = "cloud.gardener.ocm/type"

	OCITAG_ANNOTATION = "org.opencontainers.image.ref.name"
)

func RetrieveMainArtefact(m map[string]string) string {
	f, ok := m[MAINARTEFACT_ANNOTATION]
	if ok {
		return f
	}
	return m[LEGACY_MAINARTEFACT_ANNOTATION]
}

func RetrieveTags(m map[string]string) string {
	f, ok := m[TAGS_ANNOTATION]
	if ok {
		return f
	}
	return m[LEGACY_TAGS_ANNOTATION]
}

func RetrieveType(m map[string]string) string {
	f, ok := m[TYPE_ANNOTATION]
	if ok {
		return f
	}
	return m[LEGACY_TYPE_ANNOTATION]
}

// ArtefactSet provides an artefact set view on the artefact set implementation.
// Every ArtefactSet is separated closable. If the last view is closed
// the implementation is released.
type ArtefactSet struct {
	*artefactSetImpl // provide the artefact set interface
}

// implemented by view
// the rest is directly taken from the artefact set implementation

func (s *ArtefactSet) Close() error {
	return s.view.Close()
}

func (s *ArtefactSet) IsClosed() bool {
	return s.view.IsClosed()
}

////////////////////////////////////////////////////////////////////////////////

type artefactSetImpl struct {
	view support.ArtefactSetContainer
	impl support.ArtefactSetContainerImpl
	base *FileSystemBlobAccess
	*support.ArtefactSetAccess
}

var (
	_ cpi.ArtefactSink    = (*ArtefactSet)(nil)
	_ cpi.NamespaceAccess = (*ArtefactSet)(nil)
)

// New returns a new representation based element.
func New(acc accessobj.AccessMode, fs vfs.FileSystem, setup accessobj.Setup, closer accessobj.Closer, mode vfs.FileMode, formatVersion string) (*ArtefactSet, error) {
	return _Wrap(accessobj.NewAccessObject(NewAccessObjectInfo(formatVersion), acc, fs, setup, closer, mode))
}

func _Wrap(obj *accessobj.AccessObject, err error) (*ArtefactSet, error) {
	if err != nil {
		return nil, err
	}
	s := &artefactSetImpl{
		base: NewFileSystemBlobAccess(obj),
	}
	s.ArtefactSetAccess = support.NewArtefactSetAccess(s)
	s.view, s.impl = support.NewArtefactSetContainer(s)
	return &ArtefactSet{s}, nil
}

func (a *artefactSetImpl) GetNamespace() string {
	return ""
}

func (a *artefactSetImpl) Annotate(name string, value string) {
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

func (a *artefactSetImpl) AddTags(digest digest.Digest, tags ...string) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	if len(tags) == 0 {
		return nil
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
			cur := RetrieveTags(e.Annotations)
			if cur != "" {
				cur = strings.Join(append([]string{cur}, tags...), ",")
			} else {
				cur = strings.Join(tags, ",")
			}
			e.Annotations[TAGS_ANNOTATION] = cur
			e.Annotations[LEGACY_TAGS_ANNOTATION] = cur
			if a.base.FileSystemBlobAccess.Access().GetInfo().GetDescriptorFileName() == OCIArtefactSetDescriptorFileName {
				e.Annotations[OCITAG_ANNOTATION] = tags[0]
			}
			return nil
		}
	}
	return errors.ErrUnknown(cpi.KIND_OCIARTEFACT, digest.String())
}

////////////////////////////////////////////////////////////////////////////////
// forward

func (a *artefactSetImpl) IsReadOnly() bool {
	return a.base.IsReadOnly()
}

func (a *artefactSetImpl) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	return a.base.Write(path, mode, opts...)
}

func (a *artefactSetImpl) Update() error {
	return a.base.Update()
}

func (a *artefactSetImpl) Close() error {
	return a.base.Close()
}

func (a *artefactSetImpl) IsClosed() bool {
	return a.base.IsClosed()
}

// GetIndex returns the index of the included artefacts
// (image manifests and image indices)
// The manifst entries may describe dedicated tags
// to use for the dedicated artefact as annotation
// with the key TAGS_ANNOTATION.
func (a *artefactSetImpl) GetIndex() *artdesc.Index {
	if a.IsReadOnly() {
		return a.base.GetState().GetOriginalState().(*artdesc.Index)
	}
	return a.base.GetState().GetState().(*artdesc.Index)
}

// GetMain returns the digest of the main artefact
// described by this artefact set.
// There might be more, if the main artefact is an index.
func (a *artefactSetImpl) GetMain() digest.Digest {
	idx := a.GetIndex()
	if idx.Annotations == nil {
		return ""
	}
	return digest.Digest(RetrieveMainArtefact(idx.Annotations))
}

func (a *artefactSetImpl) GetBlobDescriptor(digest digest.Digest) *cpi.Descriptor {
	return a.GetIndex().GetBlobDescriptor(digest)
}

func (a *artefactSetImpl) GetBlobData(digest digest.Digest) (int64, cpi.DataAccess, error) {
	return a.base.GetBlobData(digest)
}

func (a *artefactSetImpl) AddBlob(blob cpi.BlobAccess) error {
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

func (a *artefactSetImpl) ListTags() ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Annotations != nil {
			if tags := RetrieveTags(a.Annotations); tags != "" {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *artefactSetImpl) GetTags(digest digest.Digest) ([]string, error) {
	result := []string{}
	for _, a := range a.GetIndex().Manifests {
		if a.Digest == digest && a.Annotations != nil {
			if tags := RetrieveTags(a.Annotations); tags != "" {
				result = append(result, strings.Split(tags, ",")...)
			}
		}
	}
	return result, nil
}

func (a *artefactSetImpl) HasArtefact(ref string) (bool, error) {
	if a.IsClosed() {
		return false, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.hasArtefact(ref)
}

func (a *artefactSetImpl) GetArtefact(ref string) (cpi.ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	a.base.Lock()
	defer a.base.Unlock()
	return a.getArtefact(ref)
}

func (a *artefactSetImpl) matcher(ref string) func(d *artdesc.Descriptor) bool {
	if ok, digest := artdesc.IsDigest(ref); ok {
		return func(desc *artdesc.Descriptor) bool {
			return desc.Digest == digest
		}
	}
	return func(d *artdesc.Descriptor) bool {
		if d.Annotations == nil {
			return false
		}
		for _, tag := range strings.Split(RetrieveTags(d.Annotations), ",") {
			if tag == ref {
				return true
			}
		}
		return false
	}
}

func (a *artefactSetImpl) hasArtefact(ref string) (bool, error) {
	idx := a.GetIndex()
	match := a.matcher(ref)
	for i := range idx.Manifests {
		if match(&idx.Manifests[i]) {
			return true, nil
		}
	}
	return false, nil
}

func (a *artefactSetImpl) getArtefact(ref string) (cpi.ArtefactAccess, error) {
	idx := a.GetIndex()
	match := a.matcher(ref)
	for i, e := range idx.Manifests {
		if match(&idx.Manifests[i]) {
			return a.base.GetArtefact(a.impl, e.Digest)
		}
	}
	return nil, errors.ErrUnknown(cpi.KIND_OCIARTEFACT, ref)
}

func (a *artefactSetImpl) AnnotateArtefact(digest digest.Digest, name, value string) error {
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

func (a *artefactSetImpl) AddArtefact(artefact cpi.Artefact, tags ...string) (access accessio.BlobAccess, err error) {
	blob, err := a.AddPlatformArtefact(artefact, nil)
	if err != nil {
		return nil, err
	}
	return blob, a.AddTags(blob.Digest(), tags...)
}

func (a *artefactSetImpl) AddPlatformArtefact(artefact cpi.Artefact, platform *artdesc.Platform) (access accessio.BlobAccess, err error) {
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

func (a *artefactSetImpl) NewArtefact(artefact ...*artdesc.Artefact) (cpi.ArtefactAccess, error) {
	if a.IsClosed() {
		return nil, accessio.ErrClosed
	}
	if a.IsReadOnly() {
		return nil, accessio.ErrReadOnly
	}
	return support.NewArtefact(a.impl, artefact...)
}
