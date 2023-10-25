// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"io"
	"sync"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/utils"
)

////////////////////////////////////////////////////////////////////////////////

type DefaultAccessMethod struct {
	lock   sync.Mutex
	access blobaccess.BlobAccess

	factory BlobAccessFactory
	comp    ComponentVersionAccess
	spec    AccessSpec
	mime    string
	digest  digest.Digest
	local   bool
}

var (
	_ AccessMethod            = (*DefaultAccessMethod)(nil)
	_ blobaccess.DigestSource = (*DefaultAccessMethod)(nil)
)

type BlobAccessFactory func() (BlobAccess, error)

func NewDefaultMethod(c ComponentVersionAccess, a AccessSpec, digest digest.Digest, mime string, fac BlobAccessFactory, local ...bool) AccessMethod {
	return &DefaultAccessMethod{
		spec:    a,
		comp:    c,
		mime:    mime,
		digest:  digest,
		factory: fac,
		local:   utils.Optional(local...),
	}
}

func NewDefaultMethodForBlobAccess(c ComponentVersionAccess, a AccessSpec, digest digest.Digest, blob blobaccess.BlobAccess, local ...bool) (AccessMethod, error) {
	blob, err := blob.Dup()
	if err != nil {
		return nil, err
	}
	return &DefaultAccessMethod{
		spec:    a,
		access:  blob,
		comp:    c,
		mime:    blob.MimeType(),
		digest:  digest,
		factory: nil,
		local:   utils.Optional(local...),
	}, nil
}

func (m *DefaultAccessMethod) getAccess() (blobaccess.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.access == nil {
		acc, err := m.factory()
		if err != nil {
			return nil, err
		}
		m.access = acc
	}
	return m.access, nil
}

func (m *DefaultAccessMethod) Digest() digest.Digest {
	return m.digest
}

func (m *DefaultAccessMethod) IsLocal() bool {
	return m.local
}

func (m *DefaultAccessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *DefaultAccessMethod) AccessSpec() AccessSpec {
	return m.spec
}

func (m *DefaultAccessMethod) Get() ([]byte, error) {
	a, err := m.getAccess()
	if err != nil {
		return nil, err
	}
	return a.Get()
}

func (m *DefaultAccessMethod) Reader() (io.ReadCloser, error) {
	a, err := m.getAccess()
	if err != nil {
		return nil, err
	}
	return a.Reader()
}

func (m *DefaultAccessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.access != nil {
		return m.access.Close()
	}
	return nil
}

func (m *DefaultAccessMethod) MimeType() string {
	return m.mime
}
