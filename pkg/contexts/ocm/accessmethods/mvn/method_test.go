package mvn_test

import (
	"crypto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/iotools"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const (
	mvnPATH  = "/testdata/.m2/repository"
	FAILPATH = "/testdata"
)

var _ = Describe("Method", func() {
	var env *Builder
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		env = NewBuilder(TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get packaging", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		meta, err := acc.GetPackageMeta(ocm.DefaultContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(meta.Packaging).To(Equal("pom"))
		Expect(meta.Hash).To(Equal("34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
		Expect(meta.HashType).To(Equal(crypto.SHA1))
		Expect(meta.Asc).To(ContainSubstring("-----BEGIN PGP SIGNATURE-----"))
	})

	It("accesses artifact", func() {
		acc := mvn.New("file://"+mvnPATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		Expect(m.MimeType()).To(Equal(mime.MIME_JAR)) // FIXME what about POM?

		r := Must(m.Reader())
		defer r.Close()
		dr := iotools.NewDigestReaderWithHash(crypto.SHA1, r)
		for {
			var buf [8096]byte
			_, err := dr.Read(buf[:])
			if err != nil {
				break
			}
		}
		Expect(dr.Size()).To(Equal(int64(7153)))
		Expect(dr.Digest().String()).To(Equal("SHA-1:34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
	})

	It("detects digests mismatch", func() {
		acc := mvn.New("file://"+FAILPATH, "fail", "repository", "42")

		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		_, err := m.Reader()
		Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7")))
	})
})
