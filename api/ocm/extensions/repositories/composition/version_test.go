package composition_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	me "ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/blobaccess/bpi"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/refmgmt"
)

var _ = Describe("version", func() {
	ctx := ocm.DefaultContext()

	It("handles anonymous version", func() {
		finalize := finalizer.Finalizer{}
		defer Defer(finalize.Finalize)

		nested := finalize.Nested()

		// compose new version
		cv := me.NewComponentVersion(ctx, COMPONENT, VERSION)
		cv.GetDescriptor().Provider.Name = "acme.org"
		finalize.Close(cv, "composed version")

		// wrap a non-closer access into a ref counting access to check cleanup
		blob := bpi.NewBlobAccessForBase(blobaccess.ForString(mime.MIME_TEXT, "testdata"))
		nested.Close(blob, "blob")
		MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blob, "", nil))

		// add version to repository
		repo1 := me.NewRepository(ctx)
		finalize.Close(repo1, "target repo1")
		c := Must(repo1.LookupComponent(COMPONENT))
		finalize.Close(c, "src comp")
		MustBeSuccessful(c.AddVersion(cv))
		MustBeSuccessful(nested.Finalize())

		// check result
		cv = Must(c.LookupVersion(VERSION))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))
		nested.Close(cv, "query")
		rs := Must(cv.SelectResources(selectors.Name("test")))
		Expect(len(rs)).To(Equal(1))
		data := Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))

		// add this version again
		repo2 := me.NewRepository(ctx)
		finalize.Close(repo2, "target repo2")
		MustBeSuccessful(repo2.AddComponentVersion(cv))
		Expect(refmgmt.ReferenceCount(cv)).To(Equal(1))
		MustBeSuccessful(nested.Finalize())

		// check result
		cv = Must(repo2.LookupComponentVersion(COMPONENT, VERSION))
		finalize.Close(cv, "query")
		rs = Must(cv.SelectResources(selectors.Name("test")))
		Expect(len(rs)).To(Equal(1))
		data = Must(ocmutils.GetResourceData(rs[0]))
		Expect(string(data)).To(Equal("testdata"))
	})
})
