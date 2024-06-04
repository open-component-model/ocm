package mavenblob_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/maven/maventest"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactblob/mavenblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/maven"
)

const (
	MAVEN_PATH            = "/testdata/.m2/repository"
	FAIL_PATH             = "/testdata/.m2/fail"
	MAVEN_CENTRAL_ADDRESS = "repo.maven.apache.org:443"
	MAVEN_CENTRAL         = "https://repo.maven.apache.org/maven2/"
	MAVEN_GROUP_ID        = "maven"
	MAVEN_ARTIFACT_ID     = "maven"
	MAVEN_VERSION         = "1.1"
)

var _ = Describe("blobaccess for maven", func() {

	Context("maven filesystem repository", func() {
		var env *Builder
		var repo *maven.Repository

		BeforeEach(func() {
			env = NewBuilder(maventest.TestData())
			repo = maven.NewFileRepository(MAVEN_PATH, env.FileSystem())
		})

		AfterEach(func() {
			MustBeSuccessful(env.Cleanup())
		})

		It("blobaccess for a single file with classifier and extension", func() {
			cv := composition.NewComponentVersion(env.OCMContext(), "acme.org/test", "1.0.0")
			defer Close(cv)

			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0",
				maven.WithClassifier("random-content"), maven.WithExtension("json"))

			a := me.ResourceAccessForMavenCoords(env.OCMContext(), Must(elements.ResourceMeta("mavenblob", resourcetypes.OCM_JSON, elements.WithLocalRelation())), repo, coords, me.WithCachingFileSystem(env.FileSystem()))
			Expect(a.ReferenceHint()).To(Equal(""))
			b := Must(a.BlobAccess())
			defer Close(b)
			Expect(string(Must(b.Get()))).To(Equal(`{"some": "test content"}`))

			MustBeSuccessful(cv.SetResourceAccess(a))
			r := Must(cv.GetResourceByIndex(0))
			m := Must(r.AccessMethod())
			defer Close(m)
			Expect(string(Must(m.Get()))).To(Equal(`{"some": "test content"}`))
		})

		It("blobaccess for package", func() {
			cv := composition.NewComponentVersion(env.OCMContext(), "acme.org/test", "1.0.0")
			defer Close(cv)

			coords := maven.NewCoordinates("com.sap.cloud.sdk", "sdk-modules-bom", "5.7.0")

			a := me.ResourceAccessForMavenCoords(env.OCMContext(), Must(elements.ResourceMeta("mavenblob", resourcetypes.OCM_JSON, elements.WithLocalRelation())), repo, coords, me.WithCachingFileSystem(env.FileSystem()))
			Expect(a.ReferenceHint()).To(Equal(coords.GAV()))
		})
	})
})
