// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericocireg

import (
	"io"
	"sync"

	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

type localBlobAccessMethod struct {
	lock      sync.Mutex
	data      blobaccess.DataAccess
	spec      *localblob.AccessSpec
	ref       refmgmt.Allocatable
	namespace oci.NamespaceAccess
	artifact  oci.ArtifactAccess
}

var _ accspeccpi.AccessMethodImpl = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, ns oci.NamespaceAccess, art oci.ArtifactAccess, ref refmgmt.Allocatable) (accspeccpi.AccessMethod, error) {
	return accspeccpi.AccessMethodForImplementation(newLocalBlobAccessMethodImpl(a, ns, art, ref))
}

func newLocalBlobAccessMethodImpl(a *localblob.AccessSpec, ns oci.NamespaceAccess, art oci.ArtifactAccess, ref refmgmt.Allocatable) (*localBlobAccessMethod, error) {
	err := ref.Ref()
	if err != nil {
		return nil, err
	}
	return &localBlobAccessMethod{
		spec:      a,
		ref:       ref,
		namespace: ns,
		artifact:  art,
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
	if artdesc.IsOCIMediaType(m.spec.MediaType) {
		// may be we should always store the blob, additionally to the
		// exploded form to make things easier.

		if m.spec.LocalReference == "" {
			// TODO: synthesize the artifact blob
			return nil, errors.ErrNotImplemented("artifact blob synthesis")
		}
	}
	_, data, err := m.namespace.GetBlobData(digest.Digest(m.spec.LocalReference))
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

func (m *localBlobAccessMethod) Get() ([]byte, error) {
	return blobaccess.BlobData(m.getBlob())
}

func (m *localBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
