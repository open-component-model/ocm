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
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type localOCIBlobAccessMethod struct {
	lock   sync.Mutex
	data   accessio.DataAccess
	spec   *localociblob.AccessSpec
	access oci.NamespaceAccess
}

var _ cpi.AccessMethod = (*localOCIBlobAccessMethod)(nil)

func newLocalOCIBlobAccessMethod(a *localociblob.AccessSpec, access oci.NamespaceAccess) (cpi.AccessMethod, error) {
	return &localOCIBlobAccessMethod{
		spec:   a,
		access: access,
	}, nil
}

func (m *localOCIBlobAccessMethod) GetKind() string {
	return localociblob.Type
}

func (m *localOCIBlobAccessMethod) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		tmp := m.data
		m.data = nil
		return tmp.Close()
	}
	return nil
}

func (m *localOCIBlobAccessMethod) getBlob() (cpi.DataAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.data != nil {
		return m.data, nil
	}
	data, err := m.access.GetBlobData(m.spec.Digest)
	if err != nil {
		return nil, err
	}
	m.data = data
	return m.data, err
}

func (m *localOCIBlobAccessMethod) Reader() (io.ReadCloser, error) {
	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

func (m *localOCIBlobAccessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.getBlob())
}

func (m *localOCIBlobAccessMethod) MimeType() string {
	return m.spec.GetMimeType()
}
