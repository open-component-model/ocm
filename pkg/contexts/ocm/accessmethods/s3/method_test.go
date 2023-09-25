// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package s3_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio/downloader"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/tmpcache"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/s3"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/s3/identity"
	"github.com/open-component-model/ocm/pkg/generics"
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

func checkMarshal(spec *s3.AccessSpec, typ string, fmt string) {
	if typ != "" {
		spec.SetType(typ)
	}
	data := MustWithOffset(1, Calling(json.Marshal(spec)))
	ExpectWithOffset(1, string(data)).To(Equal(fmt))

	n := MustWithOffset(1, Calling(ocm.DefaultContext().AccessSpecForConfig(data, nil)))
	Expect(reflect.TypeOf(n)).To(Equal(reflect.TypeOf(spec)))
	Expect(n.GetType()).To(Equal(generics.Conditional(typ == "", s3.Type, typ)))
	data2 := Must(json.Marshal(n))
	ExpectWithOffset(1, string(data2)).To(StringEqualWithContext(string(data)))
}

func checkDecode(spec *s3.AccessSpec, typ string, fmt string) {
	if typ != "" {
		spec.SetType(typ)
	}
	data := MustWithOffset(1, Calling(json.Marshal(spec)))

	n := MustWithOffset(1, Calling(s3.Versions().Decode([]byte(fmt), nil)))
	Expect(reflect.TypeOf(n)).To(Equal(reflect.TypeOf(spec)))

	data2 := Must(json.Marshal(n))
	ExpectWithOffset(1, string(data2)).To(StringEqualWithContext(string(data)))
}

var _ = Describe("Method", func() {
	Context("specification", func() {
		var spec *s3.AccessSpec

		BeforeEach(func() {
			spec = s3.New(
				"region",
				"bucket",
				"key",
				"version",
				"tar/gz",
			)
		})

		It("serializes", func() {
			checkMarshal(spec, "", "{\"type\":\"s3\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkMarshal(spec, s3.TypeV1, "{\"type\":\"s3/v1\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkMarshal(spec, s3.TypeV2, "{\"type\":\"s3/v2\",\"region\":\"region\",\"bucketName\":\"bucket\",\"objectKey\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkMarshal(spec, s3.LegacyType, "{\"type\":\"S3\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkMarshal(spec, s3.LegacyTypeV1, "{\"type\":\"S3/v1\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
		})

		It("deserializes versioned", func() {
			checkDecode(spec, s3.TypeV1, "{\"type\":\"s3/v1\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkDecode(spec, s3.TypeV2, "{\"type\":\"s3/v2\",\"region\":\"region\",\"bucketName\":\"bucket\",\"objectKey\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")

			checkDecode(spec, s3.LegacyTypeV1, "{\"type\":\"S3/v1\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkDecode(spec, s3.LegacyTypeV2, "{\"type\":\"S3/v2\",\"region\":\"region\",\"bucketName\":\"bucket\",\"objectKey\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
		})

		It("deserializes anonymous", func() {
			checkDecode(spec, s3.Type, "{\"type\":\"s3\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkDecode(spec, s3.Type, "{\"type\":\"s3\",\"region\":\"region\",\"bucketName\":\"bucket\",\"objectKey\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")

			checkDecode(spec, s3.LegacyType, "{\"type\":\"S3\",\"region\":\"region\",\"bucket\":\"bucket\",\"key\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
			checkDecode(spec, s3.LegacyType, "{\"type\":\"S3\",\"region\":\"region\",\"bucketName\":\"bucket\",\"objectKey\":\"key\",\"version\":\"version\",\"mediaType\":\"tar/gz\"}")
		})
	})

	Context("accessmethod", func() {
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
			env = NewBuilder()
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
			mcc = ocm.New(datacontext.MODE_INITIAL)
			mcc.CredentialsContext().SetCredentialsForConsumer(credentials.ConsumerIdentity{credentials.ID_TYPE: identity.CONSUMER_TYPE}, credentials.DirectCredentials{
				"accessKeyID":  "accessKeyID",
				"accessSecret": "accessSecret",
			})
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
})

type mockComponentVersionAccess struct {
	ocm.ComponentVersionAccess
	context ocm.Context
}

func (m *mockComponentVersionAccess) GetContext() ocm.Context {
	return m.context
}
