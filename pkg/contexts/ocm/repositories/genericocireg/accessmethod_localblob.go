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

package genericocireg

import (
	"io"
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/opencontainers/go-digest"
)

type localBlobAccessMethod struct {
	lock   sync.Mutex
	data   accessio.DataAccess
	ctx    cpi.Context
	spec   *localblob.AccessSpec
	access oci.NamespaceAccess
}

var _ cpi.AccessMethod = (*localBlobAccessMethod)(nil)

func newLocalBlobAccessMethod(a *localblob.AccessSpec, access oci.NamespaceAccess, ctx cpi.Context) (cpi.AccessMethod, error) {
	return &localBlobAccessMethod{
		ctx:    ctx,
		spec:   a,
		access: access,
	}, nil
}

func (m *localBlobAccessMethod) GetKind() string {
	return localblob.Type
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
	if artdesc.IsOCIMediaType(m.spec.MediaType) {

		// may be we should always store the blob, additionally to the
		// exploded form to make things easier.

		if m.spec.LocalReference == "" {
			// TODO: synthesize the artefact blob
			return nil, errors.ErrNotImplemented("artefact blob synthesis")
		}
	}
	_, data, err := m.access.GetBlobData(digest.Digest(m.spec.LocalReference))
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
	return accessio.BlobData(m.getBlob())
}

func (m *localBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
