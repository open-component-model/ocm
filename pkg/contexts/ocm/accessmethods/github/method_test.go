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

	_ "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"

	"k8s.io/apimachinery/pkg/util/sets"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
)

const doPrivate = false

type mockDownloader struct {
	expected        []byte
	shouldMatchLink string
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}

}

func (m *mockDownloader) Download(link string) ([]byte, error) {
	if link != m.shouldMatchLink {
		return nil, fmt.Errorf("link mismatch; got: %s want: %s", link, m.shouldMatchLink)
	}

	return m.expected, nil
}

func Configure(ctx ocm.Context) {
	data, err := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ocmconfig"))
	if err != nil {
		return
	}
	_, err = ctx.ConfigContext().ApplyData(data, nil, ".ocmconfig")
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

}

var _ = Describe("Method", func() {
	var (
		ctx                 ocm.Context
		expectedBlobContent []byte
		err                 error
		testClient          *http.Client
		defaultLink         string
		accessSpec          *me.AccessSpec
	)

	BeforeEach(func() {
		ctx = ocm.New()
		expectedBlobContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
		defaultLink = "https://github.com/test/test/sha?token=token"

		testClient = NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 302,
				Status:     http.StatusText(http.StatusFound),
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				// Must be set to non-nil value or it panics
				Header: http.Header{
					"Location": []string{defaultLink},
				},
			}
		})
		accessSpec = me.New(
			"hostname",
			1234,
			"repo",
			"owner",
			"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
			me.WithClient(testClient),
			me.WithDownloader(&mockDownloader{
				expected:        expectedBlobContent,
				shouldMatchLink: defaultLink,
			}),
		)
	})

	It("downloads public spiff commit", func() {
		spec := me.New("github.com", 0, "spiff", "mandelsoft", "25d9a3f0031c0b42e9ef7ab0117c35378040ef82")

		m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(content)).To(Equal(281655))
	})

	if doPrivate {
		Context("private access", func() {
			It("downloads private commit", func() {
				Configure(ctx)

				spec := me.New("github.com", 0, "cnudie-pause", "mandelsoft", "76eaae596ba24e401240654c4ad19ae66ba1e1a2")

				m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
				Expect(err).ToNot(HaveOccurred())
				content, err := m.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(content)).To(Equal(3764))
			})

			It("downloads enterprise commit", func() {
				Configure(ctx)

				spec := me.New("github.tools.sap", 0, "dummy", "D021770", "d17e2c594f0ab71f2c0f050b9d7fb485af4d6850")

				m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: ctx})
				Expect(err).ToNot(HaveOccurred())
				content, err := m.Get()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(content)).To(Equal(284))
			})
		})
	}

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
				1234,
				"repo",
				"owner",
				"not-a-sha",
				me.WithClient(testClient),
				me.WithDownloader(&mockDownloader{
					expected:        expectedBlobContent,
					shouldMatchLink: defaultLink,
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
				1234,
				"repo",
				"owner",
				"refs/heads/veryinteresting_branch_namess",
				me.WithClient(testClient),
				me.WithDownloader(&mockDownloader{
					expected:        expectedBlobContent,
					shouldMatchLink: defaultLink,
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
		It("can use those to access private repos", func() {
			called := false
			mcc := &mockCredContext{
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
			mcc := &mockCredContext{
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
})

type mockComponentVersionAccess struct {
	ocm.ComponentVersionAccess
	credContext ocm.Context
}

func (m *mockComponentVersionAccess) GetContext() ocm.Context {
	return m.credContext
}

type mockCredContext struct {
	ocm.Context
	creds credentials.Context
}

func (m *mockCredContext) CredentialsContext() credentials.Context {
	return m.creds
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
