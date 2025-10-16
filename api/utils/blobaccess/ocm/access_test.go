package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/selectors/rscsel"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/ocm"
	me "ocm.software/ocm/api/utils/blobaccess/ocm"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH  = "/arch.ctf"
	COMP1 = "acme.org/test1"
	COMP2 = "acme.org/test2"
	VERS  = "v1"
)

var _ = Describe("blobaccess for ocm", func() {
	Context("maven filesystem repository", func() {
		var env *Builder

		BeforeEach(func() {
			env = NewBuilder()

			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP1, VERS, func() {
					env.Resource("test", VERS, resourcetypes.PLAIN_TEXT, v1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "test data")
					})
				})

				env.ComponentVersion(COMP2, VERS, func() {
					env.Reference("ref", COMP1, VERS)
				})
			})
		})

		AfterEach(func() {
			MustBeSuccessful(env.Cleanup())
		})

		It("blobaccess for selector", func() {
			b := Must(me.BlobAccess(ocm.ByRepositorySpecAndName(env.OCMContext(), Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, accessio.PathFileSystem(env.FileSystem()))), COMP1, VERS),
				ocm.ByResourceSelector(rscsel.Name("test")),
			))
			defer Close(b, "blobaccess")

			Expect(string(Must(b.Get()))).To(Equal("test data"))
		})

		It("blobaccess for ref path", func() {
			b := Must(me.BlobAccess(ocm.ByRepositorySpecAndName(env.OCMContext(), Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, accessio.PathFileSystem(env.FileSystem()))), COMP2, VERS),
				ocm.ByResourcePath(v1.NewIdentity("test"), v1.NewIdentity("ref")),
			))
			defer Close(b, "blobaccess")

			Expect(string(Must(b.Get()))).To(Equal("test data"))
		})
	})
})
