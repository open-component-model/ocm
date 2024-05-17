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
	FAILPATH = "/testdata/fail"
)

var _ = Describe("local accessmethods.mvn.AccessSpec tests", func() {
	var env *Builder
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		env = NewBuilder(TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses local artifact", func() {
		acc := mvn.New("file://"+mvnPATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		Expect(m.MimeType()).To(Equal(mime.MIME_TGZ))
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
		Expect(dr.Size()).To(Equal(int64(1109)))
		Expect(dr.Digest().String()).To(Equal("SHA-1:4ee125ffe4f7690588833f1217a13cc741e4df5f"))
	})

	It("accesses local artifact with extension", func() {
		acc := mvn.New("file://"+mvnPATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", mvn.WithExtension("pom"))
		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		Expect(m.MimeType()).To(Equal(mime.MIME_XML))
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

	It("Describe", func() {
		acc := mvn.New("file://"+FAILPATH, "test", "repository", "42", mvn.WithExtension("pom"))
		Expect(acc.Describe(nil)).To(Equal("Maven (mvn) package 'test:repository:42::pom' in repository 'file:///testdata/fail' path 'test/repository/42/repository-42.pom'"))
	})

	It("detects digests mismatch", func() {
		acc := mvn.New("file://"+FAILPATH, "test", "repository", "42", mvn.WithExtension("pom"))
		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		_, err := m.Reader()
		Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7")))
	})
})
