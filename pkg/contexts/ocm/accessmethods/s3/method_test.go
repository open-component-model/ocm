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

package s3_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/downloader"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/s3"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
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

var _ = Describe("Method", func() {
	var (
		env             *Builder
		accessSpec      *s3.AccessSpec
		downloader      downloader.Downloader
		expectedContent []byte
		err             error
		mcc             ocm.Context
		fs              vfs.FileSystem
		ctx             datacontext.Context
	)
	BeforeEach(func() {
		expectedContent, err = os.ReadFile(filepath.Join("testdata", "repo.tar.gz"))
		Expect(err).ToNot(HaveOccurred())
		env = NewBuilder(NewEnvironment())
		downloader = &mockDownloader{
			expected: expectedContent,
		}
		accessSpec = s3.New(
			"region",
			"bucket",
			"key",
			"version",
			"tar/gz",
			downloader,
		)
		fs, err = osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		ctx = datacontext.New(nil)
		vfsattr.Set(ctx, fs)
		tmpcache.Set(ctx, &tmpcache.Attribute{Path: "/tmp"})
		mcc = &mockContext{
			dataContext: ctx,
			creds: &mockCredSource{
				cred: &mockCredentials{
					value: map[string]string{
						"accessKeyID":  "accessKeyID",
						"accessSecret": "accessSecret",
					},
				},
			},
		}
	})

	AfterEach(func() {
		env.Cleanup()
		vfs.Cleanup(fs)
	})
	It("downloads s3 objects", func() {
		m, err := accessSpec.AccessMethod(&mockComponentVersionAccess{context: mcc})
		Expect(err).ToNot(HaveOccurred())
		blob, err := m.Get()
		Expect(err).ToNot(HaveOccurred())
		Expect(blob).To(Equal(expectedContent))
	})
	When("the downloader fails to download the bucket object", func() {
		BeforeEach(func() {
			downloader = &mockDownloader{
				err: fmt.Errorf("object not found"),
			}
			accessSpec = s3.New(
				"region",
				"bucket",
				"key",
				"version",
				"tar/gz",
				downloader,
			)
		})
		It("errors", func() {
			m, err := accessSpec.AccessMethod(&mockComponentVersionAccess{context: mcc})
			Expect(err).ToNot(HaveOccurred())
			_, err = m.Get()
			Expect(err).To(MatchError(ContainSubstring("object not found")))
		})
	})
})

type mockComponentVersionAccess struct {
	ocm.ComponentVersionAccess
	context ocm.Context
}

func (m *mockComponentVersionAccess) GetContext() ocm.Context {
	return m.context
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
	value map[string]string
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
	return m.value[name]
}
