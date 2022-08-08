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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"unicode"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"

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

// Type is the access type of GitHub registry.
const Type = "gitHub"
const TypeV1 = Type + runtime.VersionSeparator + "v1"

const LegacyType = "github"
const LegacyTypeV1 = LegacyType + runtime.VersionSeparator + "v1"

const CONSUMER_TYPE = "Github"

const ShaLength = 40

func init() {
	cpi.RegisterAccessType(cpi.NewAccessSpecType(Type, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(TypeV1, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(LegacyType, &AccessSpec{}))
	cpi.RegisterAccessType(cpi.NewAccessSpecType(LegacyTypeV1, &AccessSpec{}))
}

func Is(spec cpi.AccessSpec) bool {
	return spec != nil && spec.GetKind() == Type || spec.GetKind() == LegacyType
}

// AccessSpec describes the access for a GitHub registry.
type AccessSpec struct {
	runtime.ObjectVersionedType `json:",inline"`

	// RepoUrl is the repository URL, with host, owner and repository
	RepoURL string `json:"repoUrl"`

	// APIHostname is an optional different hostname for accessing the github REST API
	// for enterprise installations
	APIHostname string `json:"apiHostname,omitempty"`

	// Ref
	Ref string `json:"ref,omitempty"`
	// Commit defines the hash of the commit.
	Commit string `json:"commit"`

	client     *http.Client
	downloader Downloader
}

var _ cpi.AccessSpec = (*AccessSpec)(nil)

// AccessSpecOptions defines a set of options which can be applied to the access spec.
type AccessSpecOptions func(s *AccessSpec)

// WithRef creates an access spec with a specified reference field
func WithRef(ref string) AccessSpecOptions {
	return func(s *AccessSpec) {
		s.Ref = ref
	}
}

// WithClient creates an access spec with a custom http client.
func WithClient(client *http.Client) AccessSpecOptions {
	return func(s *AccessSpec) {
		s.client = client
	}
}

// WithDownloader defines a client with a custom downloader.
func WithDownloader(downloader Downloader) AccessSpecOptions {
	return func(s *AccessSpec) {
		s.downloader = downloader
	}
}

// New creates a new GitHub registry access spec version v1
func New(hostname string, port int, repo, owner, commit string, opts ...AccessSpecOptions) *AccessSpec {
	if hostname == "" {
		hostname = "github.com"
	}
	p := ""
	if port != 0 {
		p = fmt.Sprintf(":%d", port)
	}
	url := fmt.Sprintf("%s%s/%s/%s", hostname, p, owner, repo)
	s := &AccessSpec{
		ObjectVersionedType: runtime.NewVersionedObjectType(Type),
		RepoURL:             url,
		Commit:              commit,
	}
	for _, o := range opts {
		o(s)
	}
	return s
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

// RepositoryService defines capabilities of a GitHub repository.
type RepositoryService interface {
	GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
}

type accessMethod struct {
	lock              sync.Mutex
	blob              artefactset.ArtefactBlob
	compvers          cpi.ComponentVersionAccess
	spec              *AccessSpec
	repositoryService RepositoryService
	owner             string
	repo              string
	downloader        Downloader
}

var _ cpi.AccessMethod = (*accessMethod)(nil)

func newMethod(c cpi.ComponentVersionAccess, a *AccessSpec) (*accessMethod, error) {
	unparsed := a.RepoURL

	if !strings.HasPrefix(unparsed, "https://") && !strings.HasPrefix(unparsed, "http://") {
		unparsed = "https://" + unparsed
	}
	u, err := url.Parse(unparsed)
	if err != nil {
		return nil, errors.ErrInvalidWrap(err, "repository url", a.RepoURL)
	}

	path := strings.Trim(u.Path, "/")
	pathcomps := strings.Split(path, "/")
	if len(pathcomps) != 2 {
		return nil, errors.ErrInvalid("repository path", path, a.RepoURL)
	}

	token, err := getCreds(u.Hostname(), u.Port(), path, c.GetContext().CredentialsContext())
	if err != nil {
		return nil, fmt.Errorf("failed to get creds: %w", err)
	}

	var client *github.Client

	httpclient := a.client

	if token != "" && httpclient == nil {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpclient = oauth2.NewClient(context.Background(), ts)
	}
	if u.Hostname() == "github.com" {
		client = github.NewClient(httpclient)
	} else {
		t := *u
		t.Path = ""
		if a.APIHostname != "" {
			t.Host = a.APIHostname
		}

		client, err = github.NewEnterpriseClient(t.String(), t.String(), httpclient)
		if err != nil {
			return nil, err
		}
	}

	var downloader Downloader = &HTTPDownloader{}
	if a.downloader != nil {
		downloader = a.downloader
	}
	return &accessMethod{
		spec:              a,
		compvers:          c,
		owner:             pathcomps[0],
		repo:              pathcomps[1],
		repositoryService: client.Repositories,
		downloader:        downloader,
	}, nil
}

func getCreds(hostname, port, path string, cctx credentials.Context) (string, error) {
	id := credentials.ConsumerIdentity{
		credentials.CONSUMER_ATTR_TYPE: CONSUMER_TYPE,
		identity.ID_HOSTNAME:           hostname,
	}
	if port != "" {
		id[identity.ID_PORT] = port
	}
	id[identity.ID_PATHPREFIX] = path
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
	if len(m.spec.Commit) != ShaLength {
		return nil, fmt.Errorf("commit is not a SHA")
	}
	for _, c := range m.spec.Commit {
		if !unicode.IsOneOf([]*unicode.RangeTable{unicode.Letter, unicode.Digit}, c) {
			return nil, fmt.Errorf("commit contains invalid characters for a SHA")
		}
	}

	link, resp, err := m.repositoryService.GetArchiveLink(context.Background(), m.owner, m.repo, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: m.spec.Commit,
	}, true)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("failed to close body: ", err)
		}
	}()
	return m.downloader.Download(link.String())
}
