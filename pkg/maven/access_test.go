package maven_test

import (
	"crypto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/optionutils"
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
		Expect(files).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 3
sdk-modules-bom-5.7.0-random-content.txt: 3
sdk-modules-bom-5.7.0-sources.jar: 3
sdk-modules-bom-5.7.0.jar: 3
sdk-modules-bom-5.7.0.pom: 3
`))
	})

	It("accesses local artifact file with extension", func() {
		repoUrl := "file://" + mvnPATH
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithClassifier(""), me.WithExtension("pom"))
		hash := Must(me.GetHash(coords.Url(repoUrl), nil, crypto.SHA1, env.FileSystem()))
		Expect(hash).To(Equal("34ccdeb9c008f8aaef90873fc636b09d3ae5c709"))
	})

	It("", func() {
		repoUrl := "file://" + mvnPATH
		coords := me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0", me.WithExtension("pom"))
		meta := Must(me.GetFileMeta(repoUrl, coords, "sdk-modules-bom-5.7.0.pom", crypto.SHA1, nil, env.FileSystem()))
		Expect(meta).To(YAMLEqual(`
  Hash: 34ccdeb9c008f8aaef90873fc636b09d3ae5c709
  HashType: 3
  MimeType: application/xml
  Url: file:///testdata/.m2/repository/com/sap/cloud/sdk/sdk-modules-bom/5.7.0/sdk-modules-bom-5.7.0.pom
`))
	})

	Context("filtering", func() {
		var (
			files  map[string]crypto.Hash
			coords *me.Coordinates
		)
		BeforeEach(func() {
			repoUrl := "file://" + mvnPATH
			coords = me.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")
			files = Must(me.GavFiles(repoUrl, coords, nil, env.FileSystem()))
		})

		It("filters nothing", func() {
			Expect(coords.FilterFileMap(files)).To(Equal(files))
		})
		It("filter by empty classifier", func() {
			coords.Classifier = optionutils.PointerTo("")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0.jar: 3
sdk-modules-bom-5.7.0.pom: 3
`))
		})
		It("filter by non-empty classifier", func() {
			coords.Classifier = optionutils.PointerTo("random-content")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-random-content.json: 3
sdk-modules-bom-5.7.0-random-content.txt: 3
`))
		})
		It("filter by extension", func() {
			coords.Extension = optionutils.PointerTo("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-sources.jar: 3
sdk-modules-bom-5.7.0.jar: 3
`))
		})

		It("filter by empty classifier and extension", func() {
			coords.Classifier = optionutils.PointerTo("")
			coords.Extension = optionutils.PointerTo("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0.jar: 3
`))
		})

		It("filter by non-empty classifier and extension", func() {
			coords.Classifier = optionutils.PointerTo("sources")
			coords.Extension = optionutils.PointerTo("jar")
			Expect(coords.FilterFileMap(files)).To(YAMLEqual(`
sdk-modules-bom-5.7.0-sources.jar: 3
`))
		})
	})
})
