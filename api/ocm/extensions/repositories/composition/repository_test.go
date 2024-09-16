package composition_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/vfs/pkg/memoryfs"

	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	me "ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/refmgmt"
)

const (
	COMPONENT = "acme.org/testcomp"
	VERSION   = "1.0.0"
)

var _ = Describe("repository", func() {
	ctx := ocm.DefaultContext()

	It("handles cvs", func() {
		finalize := finalizer.Finalizer{}
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		repo := me.NewRepository(ctx)
		finalize.Close(repo, "source repo")

		Expect(repo.GetSpecification().GetKind()).To(Equal(me.Type))

		c := Must(repo.LookupComponent(COMPONENT))
		finalize.Close(c, "src comp")

		cv := Must(c.NewVersion(VERSION))
		nested.Close(cv, "src vers")

		cv.GetDescriptor().Provider.Name = "acme.org"
		// wrap a non-closer access into a ref counting access to check cleanup
		blob := bpi.NewBlobAccessForBase(blobaccess.ForString(mime.MIME_TEXT, "testdata"))
		nested.Close(blob, "blob")
		MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blob, "", nil))
		MustBeSuccessful(c.AddVersion(cv))

		MustBeSuccessful(nested.Finalize())

		cv = Must(c.LookupVersion(VERSION))
		finalize.Close(cv, "query")
		rs := Must(cv.SelectResources(selectors.Name("test")))
		Expect(len(rs)).To(Equal(1))
		data := Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
	})

	It("supports env builder", func() {
		env := builder.NewBuilder(env.FileSystem(memoryfs.New(), ""))

		env.OCMCompositionRepository("test", func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider("acme.org")
					env.Resource("text", VERSION, "special", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize, "final")

		repo := me.NewRepository(env, "test")
		finalize.Close(repo, "repo")

		Expect(refmgmt.ReferenceCount(repo)).To(Equal(2))

		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		finalize.Close(cv, "vers")

		res := Must(cv.GetResource(metav1.NewIdentity("text")))

		data := Must(ocmutils.GetResourceData(res))
		Expect(string(data)).To(Equal("testdata"))
		Expect(refmgmt.ReferenceCount(repo)).To(Equal(3))

		MustBeSuccessful(finalize.Finalize())
		Expect(refmgmt.ReferenceCount(repo)).To(Equal(1))
	})

	It("readonly mode on repo", func() {
		env := builder.NewBuilder(env.FileSystem(memoryfs.New(), ""))

		env.OCMCompositionRepository("test", func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider("acme.org")
					env.Resource("text", VERSION, "special", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})

		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize, "final")

		sess := ocm.NewSession(nil)
		repo := me.NewRepository(env, "test")
		sess.AddCloser(repo)
		finalize.Close(sess, "repo")

		repo.SetReadOnly()
		Expect(repo.IsReadOnly()).To(BeTrue())

		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		cl := accessio.OnceCloser(cv)
		sess.AddCloser(cl)

		Expect(cv.IsReadOnly()).To(BeTrue())

		cv.GetDescriptor().Provider.Name = "acme.org"
		ExpectError(cl.Close()).To(MatchError(accessobj.ErrReadOnly))
	})

	It("provides early error", func() {
		repo := me.NewRepository(ctx)
		cv := me.NewComponentVersion(ctx, "a", "1.0")
		ExpectError(repo.AddComponentVersion(cv)).To(MatchError("component.name: Does not match pattern '^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$'"))
	})
})
