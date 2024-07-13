package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("add resources", func() {
	var ctx ocm.Context
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		ctx = ocm.New(datacontext.MODE_EXTENDED)
		cv = composition.NewComponentVersion(ctx, COMPONENT, VERSION)
	})

	AfterEach(func() {
		MustBeSuccessful(cv.Close())
	})

	Context("resources", func() {
		It("adds resource", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(Must(cv.GetResourcesByName("test"))[0].Meta().Digest).To(Equal(DS_TESTDATA))
		})

		It("adds duplicate resource with different version", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement))
			Expect(len(Must(cv.GetResourcesByName("test")))).To(Equal(2))
			Expect(Must(cv.GetResourcesByName("test"))[0].Meta().Digest).To(Equal(DS_TESTDATA))
			Expect(Must(cv.GetResourcesByName("test"))[1].Meta().Digest).To(Equal(DS_OTHERDATA))
		})

		It("rejects duplicate resource with same version", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement)).
				To(MatchError("unable to set resource: adding a new resource with same base identity requires different version"))
		})

		It("rejects duplicate resource with extra identity", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.LocalRelation).WithExtraIdentity("attr", "value")
			MustBeSuccessful(cv.SetResourceBlob(meta,
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(cv.SetResourceBlob(meta,
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement)).
				To(MatchError("unable to set resource: adding a new resource with same base identity requires different version"))
		})
	})

	Context("sources", func() {
		It("adds source", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(len(cv.GetDescriptor().Sources)).To(Equal(1))
		})

		It("adds duplicate source with different version", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement))
			Expect(len(Must(cv.GetSourcesByName("test")))).To(Equal(2))
		})

		It("rejects duplicate source with same version", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement)).
				To(MatchError("unable to set source: adding a new source with same base identity requires different version"))
		})

		It("rejects duplicate source with extra identity", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT).WithExtraIdentity("attr", "value")
			MustBeSuccessful(cv.SetSourceBlob(meta,
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			Expect(cv.SetSourceBlob(meta,
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement)).
				To(MatchError("unable to set source: adding a new source with same base identity requires different version"))
		})
	})
})
