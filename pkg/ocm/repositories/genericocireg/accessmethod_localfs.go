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

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/artdesc"
	"github.com/gardener/ocm/pkg/ocm/accessmethods/localblob"
	"github.com/gardener/ocm/pkg/ocm/cpi"
	"github.com/opencontainers/go-digest"
)

type localFilesystemBlobAccessMethod struct {
	spec   *localblob.AccessSpec
	access oci.NamespaceAccess
}

var _ cpi.AccessMethod = (*localFilesystemBlobAccessMethod)(nil)

func newLocalFilesystemBlobAccessMethod(a *localblob.AccessSpec, access oci.NamespaceAccess) (cpi.AccessMethod, error) {
	return &localFilesystemBlobAccessMethod{
		spec:   a,
		access: access,
	}, nil
}

func (m *localFilesystemBlobAccessMethod) GetKind() string {
	return localblob.Type
}

func (m *localFilesystemBlobAccessMethod) getBlob() (cpi.DataAccess, error) {
	if artdesc.IsOCIMediaType(m.spec.MediaType) {

		// may be we should always store the blob, additionally to the
		// exploded form to make things easier.

		if m.spec.LocalReference == "" {
			// TODO: synthesize the artefact blob
			return nil, errors.ErrNotImplemented("artefact blob synthesis")
		}
	}
	return m.access.GetBlobData(digest.Digest(m.spec.LocalReference))
}

func (m *localFilesystemBlobAccessMethod) Reader() (io.ReadCloser, error) {
	blob, err := m.getBlob()
	if err != nil {
		return nil, err
	}
	return blob.Reader()
}

func (m *localFilesystemBlobAccessMethod) Get() ([]byte, error) {
	return accessio.BlobData(m.getBlob())
}

func (m *localFilesystemBlobAccessMethod) MimeType() string {
	return m.spec.MediaType
}
