package npmblob_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm/elements"
	me "ocm.software/ocm/api/ocm/elements/artifactblob/npmblob"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	ocmutils "ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/tech/npm/npmtest"
)

var _ = Describe("blobaccess for npm", func() {
	Context("npm filesystem repository", func() {
		var env *Builder

		BeforeEach(func() {
			env = NewBuilder(npmtest.TestData())
		})

		AfterEach(func() {
			MustBeSuccessful(env.Cleanup())
		})

		It("blobaccess for package", func() {
			cv := composition.NewComponentVersion(env.OCMContext(), "acme.org/test", "1.0.0")
			defer Close(cv)

			a := me.ResourceAccess(env.OCMContext(), Must(elements.ResourceMeta("blob", resourcetypes.OCM_JSON, elements.WithLocalRelation())), "file://"+npmtest.NPMPATH, npmtest.PACKAGE, npmtest.VERSION, me.WithCachingFileSystem(env.FileSystem()))
			Expect(a.ReferenceHint()).To(Equal(npmtest.PACKAGE + ":" + npmtest.VERSION))

			Expect(len(Must(ocmutils.GetResourceData(a)))).To(Equal(npmtest.ARTIFACT_SIZE))
		})
	})
})
