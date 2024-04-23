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
	mvnPATH  = "/testdata/success"
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
		files, err := acc.GavFiles()
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(files["sdk-modules-bom-5.7.0.pom"]).To(Equal(crypto.SHA1))
	})

	It("get packaging", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "org.apache.maven", "apache-maven", "3.9.6")
		Expect(acc).ToNot(BeNil())
		Expect(acc.BaseUrl()).To(Equal("https://repo1.maven.org/maven2/org/apache/maven/apache-maven/3.9.6"))
		files, err := acc.GavFiles()
		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(HaveLen(8))

		//Expect(files[0]).To(Equal("sdk-modules-bom-5.7.0.pom"))
		Expect(files["apache-maven-3.9.6-src.zip"]).To(Equal(crypto.SHA512))
		Expect(files["apache-maven-3.9.6.pom"]).To(Equal(crypto.SHA1))
	})

	It("GetPackageMeta - com.sap.cloud.sdk", func() {
		acc := mvn.New("https://repo1.maven.org/maven2", "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")

		/*
			repos to test with:
				- https://repo1.maven.org/maven2/org/apache/maven/apache-maven/3.9.6/  // bin + tar.gz etc.
			    - https://repo1.maven.org/maven2/org/apache/commons/commons-compress/1.26.1/  // cyclonedx
				- https://repo1.maven.org/maven2/cn/afternode/commons/commons/1.6/ // gradle module!
			    - https://repo1.maven.org/maven2/com/sap/cloud/sdk/sdk-modules-bom/5.7.0/ // one single pom only!
			    - https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.1/ // jar only!
		*/

		meta, err := acc.GetPackageMeta(ocm.DefaultContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(meta.Bin).To(HavePrefix("file://"))
		Expect(meta.Bin).To(ContainSubstring("mvn-sdk-modules-bom-5.7.0-"))
		Expect(meta.Bin).To(HaveSuffix(".tar.gz"))
		Expect(meta.Hash).To(Equal("217feb1e7490015dd0a2b231b9cea45804df3d2a9b37287ac861bb45b8c0de55"))
		Expect(meta.HashType).To(Equal(crypto.SHA256))
	})

	/*/ works only internal
	It("GetPackageMeta - int.repositories.cloud.sap: hello-ocm", func() {
		acc := mvn.New("https://int.repositories.cloud.sap/artifactory/ocm-mvn-test", "open-component-model", "hello-ocm", "0.0.1")
		meta, err := acc.GetPackageMeta(ocm.DefaultContext())
		Expect(err).ToNot(HaveOccurred())
		Expect(meta.Bin).To(Equal("https://int.repositories.cloud.sap/artifactory/ocm-mvn-test/open-component-model/hello-ocm/0.0.1/hello-ocm-0.0.1.jar"))
		Expect(meta.Hash).To(Equal(""))
		Expect(meta.HashType).To(Equal(crypto.Hash(0)))
	})

	*/

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
		Expect(dr.Size()).To(Equal(int64(10)))
		Expect(dr.Digest().String()).To(Equal("SHA-1:e727ef4792a349c485d893e60874475a54f24b97"))
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

	It("detects digests mismatch", func() {
		acc := mvn.New("file://"+FAILPATH, "fail", "repository", "42", mvn.WithExtension("pom"))

		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		_, err := m.Reader()
		Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7")))
	})

	FIt("NewDigestWriterWithHash", func() {
		w := iotools.NewDigestWriterWithHash(crypto.SHA1, nil)
		Expect(w).ToNot(BeNil())
	})
})
