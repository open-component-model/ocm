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

package ociregistry

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/opencontainers/go-digest"
)

// Type is the access type of a oci registry.
const Type = "ociRegistry"
const TypeV1 = Type + runtime.VersionSeparator + "v1"

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
}

// AccessSpec describes the access for a oci registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// ImageReference is the actual reference to the oci image repository and tag.
	ImageReference string `json:"imageReference"`
}

// New creates a new oci registry access spec version v1
func New(ref string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		ImageReference:      ref,
	}
}

func (_ *AccessSpec) IsLocal(cpi.Context) bool {
	return false
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return newMethod(c, a)
}

////////////////////////////////////////////////////////////////////////////////

type accessMethod struct {
	lock sync.Mutex
	blob artefactset.ArtefactBlob
	comp cpi.ComponentVersionAccess
	spec *AccessSpec
}

var _ cpi.AccessMethod = (*accessMethod)(nil)
var _ accessio.DigestSource = (*accessMethod)(nil)

func newMethod(c cpi.ComponentVersionAccess, a *AccessSpec) (*accessMethod, error) {
	return &accessMethod{
		spec: a,
		comp: c,
	}, nil
}

func (m *accessMethod) GetKind() string {
	return Type
}

func (m *accessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		tmp := m.blob
		m.blob = nil
		return tmp.Close()
	}
	return nil
}

func (m *accessMethod) eval() (oci.Repository, *oci.RefSpec, error) {
	ref, err := oci.ParseRef(m.spec.ImageReference)
	if err != nil {
		return nil, nil, err
	}
	ocictx := m.comp.GetContext().OCIContext()
	spec := ocictx.GetAlias(ref.Host)
	if spec == nil {
		spec = ocireg.NewRepositorySpec(ref.Host)
	}
	repo, err := ocictx.RepositoryForSpec(spec)
	return repo, &ref, err
}

func (m *accessMethod) getArtefact() (oci.ArtefactAccess, error) {
	repo, ref, err := m.eval()
	if err != nil {
		return nil, err
	}
	return repo.LookupArtefact(ref.Repository, ref.Version())
}

func (m *accessMethod) Digest() digest.Digest {
	art, err := m.getArtefact()
	if err == nil {
		blob, err := art.Blob()
		if err == nil {
			return blob.Digest()
		}
	}
	return ""
}

func (m *accessMethod) Get() ([]byte, error) {
	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Get()
}

func (m *accessMethod) Reader() (io.ReadCloser, error) {
	b, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	r, err := b.Reader()
	if err != nil {
		return nil, err
	}
	//return accessio.AddCloser(r, b, "synthesized artefact"), nil
	return r, nil
}

func (m *accessMethod) MimeType() string {
	art, err := m.getArtefact()
	if err != nil {
		return ""
	}
	return art.GetDescriptor().MimeType()
}

func (m *accessMethod) getBlob() (artefactset.ArtefactBlob, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		return m.blob, nil
	}
	repo, ref, err := m.eval()
	if err != nil {
		return nil, err
	}
	ns, err := repo.LookupNamespace(ref.Repository)
	if err != nil {
		return nil, err
	}
	m.blob, err = artefactset.SynthesizeArtefactBlob(ns, ref.Version())
	if err != nil {
		return nil, err
	}
	return m.blob, nil
}
