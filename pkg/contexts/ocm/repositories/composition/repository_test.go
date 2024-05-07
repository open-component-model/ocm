package composition_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/memoryfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/blobaccess/bpi"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	me "github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/refmgmt"
)

const COMPONENT = "acme.org/testcomp"
const VERSION = "1.0.0"

var _ = Describe("repository", func() {
	var ctx = ocm.DefaultContext()

	It("handles cvs", func() {
		finalize := finalizer.Finalizer{}
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		repo := me.NewRepository(ctx)
		finalize.Close(repo, "source repo")

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
		rs := Must(cv.GetResourcesByName("test"))
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
})
