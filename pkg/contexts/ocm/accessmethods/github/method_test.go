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
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"net/http"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	ocmgithub "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
)

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

var _ = Describe("Method", func() {
	var (
		env                 *Builder
		expectedBlobContent []byte
		err                 error
		testClient          *http.Client
		defaultLink         string
		accessSpec          *ocmgithub.AccessSpec
	)

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
		expectedBlobContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
		defaultLink = "https://github.com/test/test/sha?token=token"

		testClient = NewTestClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: 302,
				Status:     http.StatusText(http.StatusFound),
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				// Must be set to non-nil value or it panics
				Header: http.Header{
					"Location": []string{defaultLink},
				},
			}
		})
		accessSpec = ocmgithub.New(
			"hostname",
			1234,
			"repo",
			"owner",
			"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
			ocmgithub.WithClient(testClient),
			ocmgithub.WithDownloader(&mockDownloader{
				expected:        expectedBlobContent,
				shouldMatchLink: defaultLink,
			}),
		)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads artifacts", func() {
		m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: env.OCMContext()})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal(expectedBlobContent))
	})

	When("the commit sha is of an invalid length", func() {
		It("errors", func() {
			accessSpec := ocmgithub.New(
				"hostname",
				1234,
				"repo",
				"owner",
				"not-a-sha",
				ocmgithub.WithClient(testClient),
				ocmgithub.WithDownloader(&mockDownloader{
					expected:        expectedBlobContent,
					shouldMatchLink: defaultLink,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: env.OCMContext()})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).To(MatchError(ContainSubstring("commit is not a SHA")))
		})
	})

	When("the commit sha is of the right length but contains invalid characters", func() {
		It("errors", func() {
			accessSpec := ocmgithub.New(
				"hostname",
				1234,
				"repo",
				"owner",
				"refs/heads/veryinteresting_branch_namess",
				ocmgithub.WithClient(testClient),
				ocmgithub.WithDownloader(&mockDownloader{
					expected:        expectedBlobContent,
					shouldMatchLink: defaultLink,
				}),
			)
			m, err := accessSpec.AccessMethod(&cpi.DummyComponentVersionAccess{Context: env.OCMContext()})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).To(MatchError(ContainSubstring("commit contains invalid characters for a SHA")))
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
