package maven_test

import (
	"crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	me "github.com/open-component-model/ocm/pkg/maven"
)

const (
	mvnPATH  = "/testdata/.m2/repository"
	FAILPATH = "/testdata/.m2/fail"
)

var _ = Describe("local accessmethods.me.AccessSpec tests", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses local artifact file", func() {
		repoUrl := "file://" + mvnPATH
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		files := Must(me.GavFiles(repoUrl, coords, nil, env.FileSystem()))
		Expect(files).To(YAMLEqual("sdk-modules-bom-5.7.0.pom: 3"))
	})

	It("accesses local artifact file with extension", func() {
		repoUrl := "file://" + mvnPATH
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", "", "pom")
		hash := Must(me.GetHash(coords.Url(repoUrl), nil, crypto.SHA1, env.FileSystem()))
		Expect(hash).To(Equal("34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
	})

	FIt("", func() {
		repoUrl := "file://" + mvnPATH
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", "", "pom")
		meta := Must(me.GetFileMeta(repoUrl, coords, "sdk-modules-bom-5.7.0.pom", crypto.SHA1, nil, env.FileSystem()))
		Expect(meta).To(YAMLEqual(`
  Hash: 34ccdeb9c008f8aaef90873fc636b09d3ae5c709
  HashType: 3
  MimeType: application/xml
  Url: file:///testdata/.m2/repository/com/sap/cloud/sdk/sdk-modules-bom/5.7.0/sdk-modules-bom-5.7.0.pom
`))
	})

	//It("accesses local artifact with extension", func() {
	//	acc := me.New("file://"+mvnPATH, "com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithExtension("pom"))
	//	m := Must(acc.AccessMethod(cv))
	//	defer m.Close()
	//	Expect(m.MimeType()).To(Equal(mime.MIME_XML))
	//	r := Must(m.Reader())
	//	defer r.Close()
	//	dr := iotools.NewDigestReaderWithHash(crypto.SHA1, r)
	//	for {
	//		var buf [8096]byte
	//		_, err := dr.Read(buf[:])
	//		if err != nil {
	//			break
	//		}
	//	}
	//	Expect(dr.Size()).To(Equal(int64(7153)))
	//	Expect(dr.Digest().String()).To(Equal("SHA-1:34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
	//})
	//
	//It("Describe", func() {
	//	acc := me.New("file://"+FAILPATH, "test", "repository", "42", me.WithExtension("pom"))
	//	Expect(acc.Describe(nil)).To(Equal("Maven (me) package 'test:repository:42::pom' in repository 'file:///testdata/fail' path 'test/repository/42/repository-42.pom'"))
	//})
	//
	//It("detects digests mismatch", func() {
	//	acc := me.New("file://"+FAILPATH, "test", "repository", "42", me.WithExtension("pom"))
	//	m := Must(acc.AccessMethod(cv))
	//	defer m.Close()
	//	_, err := m.Reader()
	//	Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7")))
	//})
})
