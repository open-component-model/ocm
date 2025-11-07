package maven_test

import (
	"crypto"
	"io"

	"github.com/mandelsoft/goutils/generics"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	me "ocm.software/ocm/api/tech/maven"
	"ocm.software/ocm/api/tech/maven/maventest"
)

const (
	MAVEN_PATH = "/testdata/.m2/repository"
	FAIL_PATH  = "/testdata/.m2/fail"
)

var _ = Describe("local accessmethods.me.AccessSpec tests", func() {
	var env *Builder
	var repo *me.Repository

	BeforeEach(func() {
		env = NewBuilder(maventest.TestData())
		repo = me.NewFileRepository(MAVEN_PATH, env.FileSystem())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses local artifact file", func() {
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
		files := Must(repo.GavFiles(coords, nil))
		Expect(files).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 3
sdk-modules-bom-5.7.0-random-content.txt: 3
sdk-modules-bom-5.7.0-sources.jar: 3
sdk-modules-bom-5.7.0.jar: 3
sdk-modules-bom-5.7.0.pom: 3
`))
	})

	It("accesses local artifact file with extension", func() {
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithClassifier(""), me.WithExtension("pom"))
		hash := Must(coords.Location(repo).GetHash(nil, crypto.SHA1))
		Expect(hash).To(Equal("34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
	})

	It("access dedicated file", func() {
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithClassifier(""), me.WithExtension("pom"))
		meta := Must(repo.GetFileMeta(coords, "sdk-modules-bom-5.7.0.pom", crypto.SHA1, nil))
		Expect(meta).To(YAMLEqual(`
  Hash: 34ccdeb9c008f8aaef90873fc636b09d3ae5c709
  HashType: 3
  MimeType: application/xml
  Location: /testdata/.m2/repository/com/sap/cloud/sdk/sdk-modules-bom/5.7.0/sdk-modules-bom-5.7.0.pom
`))
	})

	Context("filtering", func() {
		var (
			files  map[string]crypto.Hash
			coords *me.Coordinates
		)
		BeforeEach(func() {
			coords = me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
			files = Must(repo.GavFiles(coords, nil))
		})

		It("filters nothing", func() {
			Expect(coords.FilterFileMap(files)).To(Equal(files))
		})
		It("filter by empty classifier", func() {
			coords.Classifier = generics.Pointer("")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0.jar: 3
sdk-modules-bom-5.7.0.pom: 3
`))
		})
		It("filter by non-empty classifier", func() {
			coords.Classifier = generics.Pointer("random-content")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 3
sdk-modules-bom-5.7.0-random-content.txt: 3
`))
		})
		It("filter by extension", func() {
			coords.Extension = generics.Pointer("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-sources.jar: 3
sdk-modules-bom-5.7.0.jar: 3
`))
		})

		It("filter by empty classifier and extension", func() {
			coords.Classifier = generics.Pointer("")
			coords.Extension = generics.Pointer("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0.jar: 3
`))
		})

		It("filter by non-empty classifier and extension", func() {
			coords.Classifier = generics.Pointer("sources")
			coords.Extension = generics.Pointer("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-sources.jar: 3
`))
		})

		It("download dedicated file", func() {
			coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithClassifier(""), me.WithExtension("pom"))
			reader := Must(repo.Download(coords, nil, true))
			data := Must(io.ReadAll(reader))
			Expect(len(data)).To(Equal(7153))
			MustBeSuccessful(reader.Close())
		})

		It("download dedicated file with filed digest verification", func() {
			coords := me.NewCoordinates("test", "repository", "42", me.WithClassifier(""), me.WithExtension("pom"))
			repo := me.NewFileRepository(FAIL_PATH, env)
			reader := Must(repo.Download(coords, nil, true))
			_ = Must(io.ReadAll(reader))
			Expect(reader.Close()).To(MatchError("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found b3242b8c31f8ce14f729b8fd132ac77bc4bc5bf7"))
		})
	})
})
