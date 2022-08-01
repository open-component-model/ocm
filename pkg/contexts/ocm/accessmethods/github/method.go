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

package github

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"

	"github.com/google/go-github/v45/github"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/identity/hostpath"
	"github.com/open-component-model/ocm/pkg/contexts/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artefactset"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
	"golang.org/x/oauth2"
)

// Type is the access type of GitHub registry.
const Type = "github"
const TypeV1 = Type + runtime.VersionSeparator + "v1"
const CONSUMER_TYPE = "github"

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
}

// AccessSpec describes the access for a GitHub registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// Hostname of the GitHub installation.
	Hostname string `json:"hostname,omitempty"`
	// Port of the GitHub installation.
	Port int `json:"port,omitempty"`
	// Repository represents the name of the organization/user under which this repo can be located.
	Repository string `json:"repository"`
	// Owner represents the organization/owner of the repository.
	Owner string `json:"owner"`
	// Commit defines the hash of the commit.
	// TODO: Define this better and add example
	// TODO: Add validation that this is really a SHA and not a ref
	Commit string `json:"commit"`
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)
var _ cpi.HintProvider = (*AccessSpec)(nil)

// New creates a new GitHub registry access spec version v1
func New(hostname string, port int, repo, owner, commit string) *AccessSpec {
	return &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		Repository:          repo,
		Owner:               owner,
		Commit:              commit,
		Hostname:            hostname,
		Port:                port,
	}
}

func (_ *AccessSpec) IsLocal(cpi.Context) bool {
	return false
}

func (a *AccessSpec) GetReferenceHint() string {
	return ""
}

func (_ *AccessSpec) GetType() string {
	return Type
}

func (a *AccessSpec) AccessMethod(c cpi.ComponentVersionAccess) (cpi.AccessMethod, error) {
	return newMethod(c, a)
}

////////////////////////////////////////////////////////////////////////////////

// Repository defines capabilities of a GitHub repository.
type Repository interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
}

type accessMethod struct {
	lock             sync.Mutex
	blob             artefactset.ArtefactBlob
	comp             cpi.ComponentVersionAccess
	spec             *AccessSpec
	repositoryClient Repository
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func newMethod(c cpi.ComponentVersionAccess, a *AccessSpec) (*accessMethod, error) {

	token, err := getCreds(a, c.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get creds: %w", err)
	}

	client := github.NewClient(nil)
	if token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(ctx, ts)

		client = github.NewClient(tc)
	}
	return &accessMethod{
		spec:             a,
		comp:             c,
		repositoryClient: client.Repositories,
	}, nil
}

func getCreds(a *AccessSpec, cctx credentials.Context) (string, error) {
	hostname := "github.com"
	if a.Hostname != "" {
		hostname = a.Hostname
	}
	id := credentials.ConsumerIdentity{
		credentials.CONSUMER_ATTR_TYPE: CONSUMER_TYPE,
		identity.ID_HOSTNAME:           hostname,
	}
	if a.Port != 0 {
		id[identity.ID_PORT] = strconv.Itoa(a.Port)
	}
	id[identity.ID_PATHPREFIX] = path.Join(a.Owner, a.Repository)
	var creds credentials.Credentials
	src, err := cctx.GetCredentialsForConsumer(id, hostpath.IdentityMatcher(CONSUMER_TYPE))
	if err != nil {
		if !errors.IsErrUnknown(err) {
			return "", err
		}
		return "", nil
	}
	if src != nil {
		creds, err = src.Credentials(cctx)
		if err != nil {
			return "", err
		}
	}
	return creds.GetProperty(credentials.ATTR_TOKEN), nil
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
	blob, err := m.downloadArchive()
	if err != nil {
		return nil, err
	}

	return accessio.BlobAccessForData(mime.MIME_TGZ, blob), nil
}

func (m *accessMethod) downloadArchive() ([]byte, error) {
	link, resp, err := m.repositoryClient.GetArchiveLink(context.Background(), m.spec.Owner, m.spec.Repository, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: m.spec.Commit,
	}, true)
	if err != nil {
		fmt.Println("err from github: ", err)
		fmt.Println("trying to read body for more information...")
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("failed to read body")
			return nil, err
		}
		fmt.Println("body: ", string(content))
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("failed to close body: ", err)
		}
	}()
	httpResp, err := http.Get(link.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			fmt.Println("failed to close body: ", err)
		}
	}()

	var blob []byte
	buf := bytes.NewBuffer(blob)
	if _, err := io.Copy(buf, httpResp.Body); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
