package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/selectors"
	. "ocm.software/ocm/api/ocm/testhelper"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

const (
	COMPONENT = "acme.org/test"
	VERSION   = "v1"
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
			Expect(Must(cv.SelectResources(selectors.Name("test")))[0].Meta().Digest).To(Equal(DS_TESTDATA))
		})

		It("replaces resource", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil))
			Expect(Must(cv.SelectResources(selectors.Name("test")))[0].Meta().Digest).To(Equal(DS_OTHERDATA))
		})

		It("replaces resource (enforced)", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement))
			Expect(Must(cv.SelectResources(selectors.Name("test")))[0].Meta().Digest).To(Equal(DS_OTHERDATA))

			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement))
			Expect(Must(cv.SelectResources(selectors.Name("test")))[0].Meta().Digest).To(Equal(DS_OTHERDATA))
		})

		It("fails replace non-existent resource)", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			Expect(cv.SetResourceBlob(meta.WithVersion("v1").WithExtraIdentity("attr", "value"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement)).To(
				MatchError("unable to set resource: element \"attr\"=\"value\",\"name\"=\"test\" not found"))
		})

		It("adds duplicate resource with different version", func() {
			meta := ocm.NewResourceMeta("test", resourcetypes.PLAIN_TEXT, metav1.ExternalRelation)
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			MustBeSuccessful(cv.SetResourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement))
			Expect(len(Must(cv.SelectResources(selectors.Name("test"))))).To(Equal(2))
			Expect(Must(cv.SelectResources(selectors.Name("test")))[0].Meta().Digest).To(Equal(DS_TESTDATA))
			Expect(Must(cv.SelectResources(selectors.Name("test")))[1].Meta().Digest).To(Equal(DS_OTHERDATA))
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

		It("replaces source", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil))
			Expect(len(Must(cv.SelectSources(selectors.Name("test"))))).To(Equal(1))
		})

		It("replaces source (enforced)", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement))
			Expect(len(Must(cv.SelectSources(selectors.Name("test"))))).To(Equal(1))

			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement))
			Expect(len(Must(cv.SelectSources(selectors.Name("test"))))).To(Equal(1))
		})

		It("fails replace non-existent source)", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))

			Expect(cv.SetSourceBlob(meta.WithVersion("v1").WithExtraIdentity("attr", "value"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.UpdateElement)).To(
				MatchError("unable to set source: element \"attr\"=\"value\",\"name\"=\"test\" not found"))
		})

		It("adds duplicate source with different version", func() {
			meta := ocm.NewSourceMeta("test", resourcetypes.PLAIN_TEXT)
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v1"),
				blobaccess.ForString(mime.MIME_TEXT, S_TESTDATA), "", nil))
			MustBeSuccessful(cv.SetSourceBlob(meta.WithVersion("v2"),
				blobaccess.ForString(mime.MIME_TEXT, S_OTHERDATA), "", nil, ocm.AppendElement))
			Expect(len(Must(cv.SelectSources(selectors.Name("test"))))).To(Equal(2))
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

	Context("references", func() {
		It("adds reference", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))
			Expect(len(cv.GetDescriptor().References)).To(Equal(1))
		})

		It("replaces reference", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))

			MustBeSuccessful(cv.SetReference(ref.WithVersion("v1")))
			Expect(len(Must(cv.SelectReferences(selectors.Name("test"))))).To(Equal(1))
		})

		It("replaces source (enforced)", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))

			MustBeSuccessful(cv.SetReference(ref.WithVersion("v2")))
			Expect(len(Must(cv.SelectReferences(selectors.Name("test"))))).To(Equal(1))
		})

		It("fails replace non-existent source)", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))

			Expect(cv.SetReference(ref.WithExtraIdentity("attr", "value"), ocm.UpdateElement)).To(
				MatchError("element \"attr\"=\"value\",\"name\"=\"test\" not found"))
		})

		It("adds duplicate reference with different version", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))
			MustBeSuccessful(cv.SetReference(ref.WithVersion("v2"), ocm.AppendElement))
			Expect(len(Must(cv.SelectReferences(selectors.Name("test"))))).To(Equal(2))
		})

		It("rejects duplicate reference with same version", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))
			Expect(cv.SetReference(ref.WithVersion("v1"), ocm.AppendElement)).
				To(MatchError("adding a new reference with same base identity requires different version"))
		})

		It("rejects duplicate reference with extra identity", func() {
			ref := ocm.NewComponentReference("test", COMPONENT+"/sub", "v1").WithExtraIdentity("attr", "value")
			MustBeSuccessful(cv.SetReference(ref, ocm.ModifyElement()))
			Expect(cv.SetReference(ref, ocm.AppendElement)).
				To(MatchError("adding a new reference with same base identity requires different version"))
		})
	})
})
