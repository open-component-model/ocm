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

package s3

import (
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

// Type is the access type of S3 registry.
const Type = "S3"
const TypeV1 = Type + runtime.VersionSeparator + "v1"
const CONSUMER_TYPE = "s3"

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
}

// AccessSpec describes the access for a GitHub registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Region needs to be set even though buckets are global.
	// We can't assume that there is a default region setting sitting somewhere.
	Region string `json:"region"`
	// Bucket where the s3 object is located.
	Bucket string `json:"bucket"`
	// Key of the object to look for. This value will be used together with Bucket and Version to form an identity.
	Key string `json:"key"`
	// Version of the object.
	// +optional
	Version string `json:"version,omitempty"`

	downloader Downloader
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)

// New creates a new GitHub registry access spec version v1
func New(region, bucket, key, version string, downloader Downloader) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Region:              region,
		Bucket:              bucket,
		Key:                 key,
		Version:             version,
		downloader:          downloader,
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
	lock         sync.Mutex
	blob         artefactset.ArtefactBlob
	comp         cpi.ComponentVersionAccess
	spec         *AccessSpec
	accessKeyID  string
	accessSecret string
	downloader   Downloader
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func newMethod(c cpi.ComponentVersionAccess, a *AccessSpec) (*accessMethod, error) {
	creds, err := getCreds(a, c.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get creds: %w", err)
	}

	if creds == nil {
		return nil, fmt.Errorf("failed to return any credentials; they MUST be provided for s3 access")
	}

	var d Downloader = &S3Downloader{}
	if a.downloader != nil {
		d = a.downloader
	}

	return &accessMethod{
		spec:         a,
		comp:         c,
		accessKeyID:  creds.GetProperty(credentials.ATTR_AWS_ACCESS_KEY_ID),
		accessSecret: creds.GetProperty(credentials.ATTR_AWS_SECRET_ACCESS_KEY),
		downloader:   d,
	}, nil
}

func getCreds(a *AccessSpec, cctx credentials.Context) (credentials.Credentials, error) {
	id := credentials.ConsumerIdentity{
		credentials.CONSUMER_ATTR_TYPE: CONSUMER_TYPE,
		identity.ID_HOSTNAME:           a.Bucket,
	}
	if a.Version != "" {
		id[identity.ID_PORT] = a.Version
	}
	id[identity.ID_PATHPREFIX] = path.Join(a.Bucket, a.Key, a.Version)
	var creds credentials.Credentials
	src, err := cctx.GetCredentialsForConsumer(id, hostpath.IdentityMatcher(CONSUMER_TYPE))
	if err != nil {
		if !errors.IsErrUnknown(err) {
			return nil, err
		}
		return nil, nil
	}
	if src != nil {
		creds, err = src.Credentials(cctx)
		if err != nil {
			return nil, err
		}
	}
	return creds, nil
}

func (m *accessMethod) GetKind() string {
	return Type
}

// Close should clean up all cached data if present.
// Exp.: Cache the blob data.
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
	return r, nil
}

func (m *accessMethod) MimeType() string {
	return mime.MIME_TGZ
}

// TODO: Implement caching based on the SHA of the blob. If it is detected that that SHA already exists
// return it. ( Use the virtual filesystem implementation so it can be in memory or via file system ).
func (m *accessMethod) getBlob() (accessio.BlobAccess, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.blob != nil {
		return m.blob, nil
	}
	blob, err := m.downloader.Download(
		m.spec.Region,
		m.spec.Bucket,
		m.spec.Key,
		m.spec.Version,
		m.accessKeyID,
		m.accessSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}

	return accessio.BlobAccessForData(mime.MIME_TGZ, blob), nil
}
