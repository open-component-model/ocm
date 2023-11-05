// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package virtual

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

type localBlobAccessMethod struct {
	lock   sync.Mutex
	data   blobaccess.DataAccess
	spec   *localblob.AccessSpec
	ref    refmgmt.Allocatable
	access VersionAccess
}

var _ accspeccpi.AccessMethodImpl = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, acc VersionAccess, ref refmgmt.Allocatable) (*localBlobAccessMethod, error) {
	err := ref.Ref()
	if err != nil {
		return nil, err
	}
	return &localBlobAccessMethod{
		spec:   a,
		ref:    ref,
		access: acc,
	}, nil
}

func (_ *localBlobAccessMethod) IsLocal() bool {
	return true
}

func (m *localBlobAccessMethod) GetKind() string {
	return m.spec.GetKind()
}

func (m *localBlobAccessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *localBlobAccessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	list := errors.ErrorList{}

	if m.data != nil {
		tmp := m.data
		m.data = nil
		list.Add(tmp.Close())
	}
	list.Add(m.ref.Unref())
	return list.Result()
}

func (m *localBlobAccessMethod) getBlob() (blobaccess.DataAccess, error) {
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
