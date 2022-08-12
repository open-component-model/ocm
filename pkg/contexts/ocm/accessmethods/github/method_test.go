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

package github_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

type mockDownloader struct {
	expected []byte
	err      error
}

func (m *mockDownloader) Download(w io.WriterAt) error {
	if _, err := w.WriteAt(m.expected, 0); err != nil {
		return fmt.Errorf("failed to write to mock writer: %w", err)
	}
	return m.err
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}

}

var _ = Describe("Method", func() {
	var (
		ctx                 ocm.Context
		expectedBlobContent []byte
		err                 error
		defaultLink         string
		accessSpec          *me.AccessSpec
		dctx                datacontext.Context
		fs                  vfs.FileSystem
		expectedURL         string
		clientFn            func(url string) *http.Client
	)

	BeforeEach(func() {
		ctx = ocm.New()
		expectedBlobContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
		defaultLink = "https://github.com/test/test/sha?token=token"
		expectedURL = "https://api.github.com/repos/test/test/tarball/7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f"

		clientFn = func(url string) *http.Client {
			return NewTestClient(func(req *http.Request) *http.Response {
				if req.URL.String() != url {
					Fail(fmt.Sprintf("failed to match url to expected url. want: %s; got: %s", expectedURL, req.URL.String()))
				}
				return &http.Response{
					StatusCode: 302,
					Status:     http.StatusText(http.StatusFound),
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					Header: http.Header{
						"Location": []string{defaultLink},
					},
				}
			})
		}

		accessSpec = me.New(
			"https://github.com/test/test",
			"",
			"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
			me.WithClient(clientFn(expectedURL)),
			me.WithDownloader(&mockDownloader{
				expected: expectedBlobContent,
			}),
		)
		fs, err = osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		dctx = datacontext.New(nil)
		vfsattr.Set(ctx, fs)
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: "/tmp"})
	})

	AfterEach(func() {
		vfs.Cleanup(fs)
	})

	It("downloads artifacts", func() {
		m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal(expectedBlobContent))
	})

	When("the commit sha is of an invalid length", func() {
		It("errors", func() {
			accessSpec := me.New(
				"hostname",
				"",
				"not-a-sha",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).To(MatchError(ContainSubstring("commit is not a SHA")))
			if m != nil {
				m.Close()
			}
		})
	})

	When("the commit sha is of the right length but contains invalid characters", func() {
		It("errors", func() {
			accessSpec := me.New(
				"hostname",
				"1234",
				"refs/heads/veryinteresting_branch_namess",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).To(MatchError(ContainSubstring("commit contains invalid characters for a SHA")))
			if m != nil {
				m.Close()
			}
		})
	})

	When("credentials are provided", func() {
		BeforeEach(func() {
			clientFn = func(url string) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if v, ok := req.Header["Authorization"]; ok {
						Expect(v).To(ContainElement("Bearer test"))
					} else {
						Fail("Authorization header not found in request")
					}
					if req.URL.String() != url {
						Fail(fmt.Sprintf("failed to match url to expected url. want: %s; got: %s", expectedURL, req.URL.String()))
					}
					return &http.Response{
						StatusCode: 302,
						Status:     http.StatusText(http.StatusFound),
						// Must be set to non-nil value or it panics
						Body: io.NopCloser(bytes.NewBufferString(`{}`)),
						Header: http.Header{
							"Location": []string{defaultLink},
						},
					}
				})
			}
			accessSpec = me.New(
				"https://github.com/test/test",
				"",
				"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
				me.WithClient(clientFn(expectedURL)),
				me.WithDownloader(&mockDownloader{
					expected: expectedBlobContent,
				}),
			)
		})
		It("can use those to access private repos", func() {
			called := false
			mcc := &mockContext{
				dataContext: dctx,
				creds: &mockCredSource{
					cred: &mockCredentials{
						value: func() string {
							called = true
							return "test"
						},
					},
				},
			}
			m, err := accessSpec.AccessMethod(&mockComponentVersionAccess{
				credContext: mcc,
			})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).ToNot(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})

	When("GetCredentialsForConsumer returns an error", func() {
		It("errors", func() {
			called := false
			mcc := &mockContext{
				creds: &mockCredSource{
					cred: &mockCredentials{
						value: func() string {
							called = true
							return "test"
						},
					},
					err: fmt.Errorf("danger will robinson"),
				},
			}
			_, err := accessSpec.AccessMethod(&mockComponentVersionAccess{
				credContext: mcc,
			})
			Expect(err).To(MatchError(ContainSubstring("danger will robinson")))
			Expect(called).To(BeFalse())
		})
	})

	When("an enterprise repo URL is provided", func() {
		It("uses that domain and includes api/v3 in the request URL", func() {
			expectedURL = "https://github.tools.sap/api/v3/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("https://github.tools.sap/test/test", "", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("hostname is different from github.com", func() {
		It("will use an enterprise client", func() {
			expectedURL = "https://custom/api/v3/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("https://github.tools.sap/test/test", "custom", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	When("repoURL doesn't have an https prefix", func() {
		It("will add one", func() {
			expectedURL = "https://api.github.com/repos/test/test/tarball/25d9a3f0031c0b42e9ef7ab0117c35378040ef82"
			spec := me.New("github.com/test/test", "", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82", me.WithClient(clientFn(expectedURL)))
			_, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

type mockComponentVersionAccess struct {
	ocm.ComponentVersionAccess
	credContext ocm.Context
}

func (m *mockComponentVersionAccess) GetContext() ocm.Context {
	return m.credContext
}

type mockContext struct {
	ocm.Context
	creds       credentials.Context
	dataContext datacontext.Context
}

func (m *mockContext) CredentialsContext() credentials.Context {
	return m.creds
}

func (m *mockContext) GetAttributes() datacontext.Attributes {
	return m.dataContext.GetAttributes()
}

type mockCredSource struct {
	credentials.Context
	cred credentials.Credentials
	err  error
}

func (m *mockCredSource) GetCredentialsForConsumer(credentials.ConsumerIdentity, ...credentials.IdentityMatcher) (credentials.CredentialsSource, error) {
	return m, m.err
}

func (m *mockCredSource) Credentials(credentials.Context, ...credentials.CredentialsSource) (credentials.Credentials, error) {
	return m.cred, nil
}

type mockCredentials struct {
	value func() string
}

func (m *mockCredentials) Credentials(context core.Context, source ...core.CredentialsSource) (core.Credentials, error) {
	panic("implement me")
}

func (m *mockCredentials) ExistsProperty(name string) bool {
	panic("implement me")
}

func (m *mockCredentials) PropertyNames() sets.String {
	panic("implement me")
}

func (m *mockCredentials) Properties() common.Properties {
	panic("implement me")
}

func (m *mockCredentials) GetProperty(name string) string {
	return m.value()
}
