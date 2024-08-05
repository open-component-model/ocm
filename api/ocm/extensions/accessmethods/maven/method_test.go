package maven_test

import (
	"crypto"
	"time"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	me "ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	"ocm.software/ocm/api/tech/maven/maventest"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
)

const (
	MAVEN_PATH            = "/testdata/.m2/repository"
	FAILPATH              = "/testdata/.m2/fail"
	MAVEN_CENTRAL         = "https://repo.maven.apache.org/maven2/"
	MAVEN_CENTRAL_ADDRESS = "repo.maven.apache.org:443"
	MAVEN_GROUP_ID        = "maven"
	MAVEN_ARTIFACT_ID     = "maven"
	MAVEN_VERSION         = "1.1"
)

var _ = Describe("local accessmethods.maven.AccessSpec tests", func() {
	var env *Builder
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		env = NewBuilder(maventest.TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses local artifact", func() {
		acc := me.New("file://"+MAVEN_PATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		m := Must(acc.AccessMethod(cv))
		defer Close(m)
		Expect(m.MimeType()).To(Equal(mime.MIME_TGZ))
		r := Must(m.Reader())
		defer Close(r)
		dr := iotools.NewDigestReaderWithHash(crypto.SHA256, r)
		li := Must(tarutils.ListArchiveContentFromReader(dr))
		Expect(li).To(ConsistOf(
			"sdk-modules-bom-5.7.0-random-content.json",
			"sdk-modules-bom-5.7.0-random-content.txt",
			"sdk-modules-bom-5.7.0-sources.jar",
			"sdk-modules-bom-5.7.0.jar",
			"sdk-modules-bom-5.7.0.pom"))
		Expect(dr.Size()).To(Equal(int64(maventest.ARTIFACT_SIZE)))
		Expect(dr.Digest().String()).To(Equal("SHA-256:" + maventest.ARTIFACT_DIGEST))
	})
	It("test empty repoUrl", func() {
		acc := me.New("", "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		ExpectError(acc.AccessMethod(cv)).ToNot(BeNil())
	})

	It("accesses local artifact with empty classifier and with extension", func() {
		acc := me.New("file://"+MAVEN_PATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithClassifier(""), me.WithExtension("pom"))
		m := Must(acc.AccessMethod(cv))
		defer Close(m)
		Expect(m.MimeType()).To(Equal(mime.MIME_XML))
		r := Must(m.Reader())
		defer Close(r)

		dr := iotools.NewDigestReaderWithHash(crypto.SHA1, r)
		for {
			var buf [8096]byte
			_, err := dr.Read(buf[:])
			if err != nil {
				break
			}
		}

		Expect(dr.Size()).To(Equal(int64(7153)))
		Expect(dr.Digest().String()).To(Equal(maventest.POM_SHA1))
	})

	It("accesses local artifact with extension", func() {
		acc := me.New("file://"+MAVEN_PATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithExtension("pom"))
		m := Must(acc.AccessMethod(cv))
		defer Close(m)
		Expect(m.MimeType()).To(Equal(mime.MIME_TGZ))
		r := Must(m.Reader())
		defer Close(r)
		dr := iotools.NewDigestReaderWithHash(crypto.SHA1, r)
		list := Must(tarutils.ListArchiveContentFromReader(dr))
		Expect(list).To(ConsistOf("sdk-modules-bom-5.7.0.pom"))

		Expect(dr.Size()).To(Equal(int64(1109)))
		Expect(dr.Digest().String()).To(Equal("SHA-1:4ee125ffe4f7690588833f1217a13cc741e4df5f"))
	})

	It("Describe", func() {
		acc := me.New("file://"+FAILPATH, "test", "repository", "42", me.WithExtension("pom"))
		Expect(acc.Describe(nil)).To(Equal("Maven package 'test:repository:42::pom' in repository 'file:///testdata/.m2/fail' path 'test/repository/42/repository-42.pom'"))
	})

	It("detects digests mismatch", func() {
		acc := me.New("file://"+FAILPATH, "test", "repository", "42", me.WithExtension("pom"))
		m := Must(acc.AccessMethod(cv))
		defer Close(m)
		_, err := m.Reader()
		Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7")))
	})

	Context("me http repository", func() {
		if PingTCPServer(MAVEN_CENTRAL_ADDRESS, time.Second) == nil {
			It("blobaccess for gav", func() {
				acc := me.New(MAVEN_CENTRAL, MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION)
				m := Must(acc.AccessMethod(cv))
				defer Close(m)
				files := Must(tarutils.ListArchiveContentFromReader(Must(m.Reader())))
				Expect(files).To(ConsistOf(
					"maven-1.1-RC1.javadoc.javadoc.jar",
					"maven-1.1-sources.jar",
					"maven-1.1.jar",
					"maven-1.1.pom",
				))
			})

			It("inexpensive id", func() {
				acc := me.New(MAVEN_CENTRAL, MAVEN_GROUP_ID, MAVEN_ARTIFACT_ID, MAVEN_VERSION, me.WithClassifier(""), me.WithExtension("pom"))
				Expect(acc.ArtifactId).To(Equal("maven"))
			})
		}
	})
})
