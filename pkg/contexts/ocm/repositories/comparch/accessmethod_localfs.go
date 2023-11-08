// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package comparch

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/support"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

////////////////////////////////////////////////////////////////////////////////

type localFilesystemBlobAccessMethod struct {
	sync.Mutex
	ref        refmgmt.Allocatable
	closed     bool
	spec       *localblob.AccessSpec
	base       support.ComponentVersionContainer
	blobAccess blobaccess.DataAccess
}

var _ accspeccpi.AccessMethodImpl = (*localFilesystemBlobAccessMethod)(nil)

func newLocalFilesystemBlobAccessMethod(a *localblob.AccessSpec, base support.ComponentVersionContainer, ref refmgmt.ExtendedAllocatable) (accspeccpi.AccessMethod, error) {
	err := ref.Ref()
	if err != nil {
		return nil, err
	}

	m, _ := accspeccpi.AccessMethodForImplementation(&localFilesystemBlobAccessMethod{
		spec: a,
		base: base,
		ref:  ref,
	}, nil)
	return m, nil
}

func (_ *localFilesystemBlobAccessMethod) IsLocal() bool {
	return true
}

func (m *localFilesystemBlobAccessMethod) AccessSpec() accspeccpi.AccessSpec {
	return m.spec
}

func (m *localFilesystemBlobAccessMethod) GetKind() string {
	return localblob.Type
}

func (m *localFilesystemBlobAccessMethod) Reader() (io.ReadCloser, error) {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return nil, accessio.ErrClosed
	}

	if m.blobAccess == nil {
		var err error
		m.blobAccess, err = m.base.GetBlobData(m.spec.LocalReference)
		if err != nil {
			return blobaccess.BlobReader(m.blobAccess, err)
		}
	}
	return blobaccess.BlobReader(m.blobAccess, nil)
}

func (m *localFilesystemBlobAccessMethod) Get() ([]byte, error) {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return nil, accessio.ErrClosed
	}

	if m.blobAccess == nil {
		var err error
		m.blobAccess, err = m.base.GetBlobData(m.spec.LocalReference)
		if err != nil {
			return blobaccess.BlobData(m.blobAccess, err)
		}
	}
	return blobaccess.BlobData(m.blobAccess, nil)
}

func (m *localFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}

func (m *localFilesystemBlobAccessMethod) Close() error {
	m.Lock()
	defer m.Unlock()

	if m.closed {
		return accessio.ErrClosed
	}

	list := errors.ErrorList{}
	if m.blobAccess != nil {
		list.Add(m.blobAccess.Close())
		m.blobAccess = nil
	}
	list.Add(m.ref.Unref())
	m.closed = true
	return list.Result()
}
