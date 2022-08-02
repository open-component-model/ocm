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
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v45/github"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ocmgithub "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/github"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
)

type mockRepoService struct {
}

func (m *mockRepoService) GetArchiveLink(ctx context.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error) {
	link, err := url.Parse("https://github.com/org/repo")
	if err != nil {
		return nil, nil, err
	}
	return link, &github.Response{
		Response: &http.Response{
			Body: io.NopCloser(strings.NewReader("body")),
		},
	}, nil
}

type mockDownloader struct {
	expected []byte
}

func (m *mockDownloader) Download(link string) ([]byte, error) {
	return m.expected, nil
}

var _ = Describe("Method", func() {
	var (
		env                 *Builder
		expectedBlobContent []byte
		err                 error
	)

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())
		expectedBlobContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads artifacts", func() {
		accessMethod := ocmgithub.New(
			"hostname",
			1234,
			"repo",
			"owner",
			"7b1445755ee2527f0bf80ef9eeb59a5d2e6e3e1f",
			&mockRepoService{},
			&mockDownloader{
				expected: expectedBlobContent,
			},
		)
		m, err := accessMethod.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
		Expect(err).ToNot(HaveOccurred())
		content, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(content).To(Equal(expectedBlobContent))
	})

	When("the commit sha is of an invalid length", func() {
		It("errors", func() {
			accessMethod := ocmgithub.New(
				"hostname",
				1234,
				"repo",
				"owner",
				"not-a-sha",
				&mockRepoService{},
				&mockDownloader{
					expected: expectedBlobContent,
				},
			)
			m, err := accessMethod.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).To(MatchError(ContainSubstring("commit is not a SHA")))
		})
	})

	When("the commit sha is of the right length but contains invalid characters", func() {
		It("errors", func() {
			accessMethod := ocmgithub.New(
				"hostname",
				1234,
				"repo",
				"owner",
				"refs/heads/veryinteresting_branch_namess",
				&mockRepoService{},
				&mockDownloader{
					expected: expectedBlobContent,
				},
			)
			m, err := accessMethod.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).To(MatchError(ContainSubstring("commit contains invalid characters for a SHA")))
		})
	})
})
