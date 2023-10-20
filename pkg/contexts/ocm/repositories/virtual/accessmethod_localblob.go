// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type localBlobAccessMethod struct {
	lock   sync.Mutex
	data   blobaccess.DataAccess
	spec   *localblob.AccessSpec
	access VersionAccess
}

var _ cpi.AccessMethod = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, acc VersionAccess) *localBlobAccessMethod {
	return &localBlobAccessMethod{
		spec:   a,
		access: acc,
	}
}

func (_ *localBlobAccessMethod) IsLocal() bool {
	return true
}

func (m *localBlobAccessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *localBlobAccessMethod) AccessSpec() cpi.AccessSpec {
	return m.spec
}

func (m *localBlobAccessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		tmp := m.data
		m.data = nil
		return tmp.Close()
	}
	return nil
}

func (m *localBlobAccessMethod) getBlob() (cpi.DataAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		return m.data, nil
	}
	data, err := m.access.GetBlob(m.spec.LocalReference)
	if err != nil {
		return nil, err
	}
	m.data = data
	return m.data, err
}

func (m *localBlobAccessMethod) Reader() (io.ReadCloser, error) {
	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

func (m *localBlobAccessMethod) Get() (data []byte, ferr error) {
	b, err := m.getBlob()
	if ferr != nil {
		return nil, err
	}
	return blobaccess.BlobData(b)
}

func (m *localBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
