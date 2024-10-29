package signing_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	rsa_pss "ocm.software/ocm/api/tech/signing/handlers/rsa-pss"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

var _ = Describe("Simple signing handlers", func() {
	ctx := ocm.DefaultContext()

	var cv ocm.ComponentVersionAccess
	var pub signutils.GenericPublicKey
	var priv signutils.GenericPrivateKey

	BeforeEach(func() {
		priv, pub = Must2(rsa.CreateKeyPair())
	})

	Context("standard", func() {
		BeforeEach(func() {
			cv = composition.NewComponentVersion(ctx, COMPONENTA, VERSION)
			MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation), blobaccess.ForString(mime.MIME_TEXT, "test data"), "", nil))
		})

		DescribeTable("rsa handlers", func(kind string) {
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv), signing.SignerByAlgo(kind)))
			Must(signing.VerifyComponentVersion(cv, "signature", signing.PublicKey("signature", pub)))
		},
			Entry("rsa", rsa.Algorithm),
			Entry("rsapss", rsa_pss.Algorithm),
		)

		It("uses verified store", func() {
			store := signing.NewLocalVerifiedStore()
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv), signing.UseVerifiedStore(store)))

			cd := store.Get(cv)
			Expect(cd).NotTo(BeNil())
			Expect(len(cd.Signatures)).To(Equal(1))
		})
	})

	Context("non-unique resources", func() {
		BeforeEach(func() {
			cv = composition.NewComponentVersion(ctx, COMPONENTA, VERSION)

			meta := ocm.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation)
			meta.Version = "v1"
			MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, "test data"), "", nil))
			meta.ExtraIdentity = map[string]string{}
			meta.Version = "v2"
			MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, "other test data"), "", nil, ocm.TargetIndex(-1)))
		})

		It("signs without modification (compatibility)", func() {
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv)))
			cd := cv.GetDescriptor()
			cd.Resources[0].ExtraIdentity = v1.Identity{}
			cd.Resources[1].ExtraIdentity = v1.Identity{}
			Expect(len(cd.Resources)).To(Equal(2))
			Expect(len(cd.Resources[0].ExtraIdentity)).To(Equal(0))
			Expect(len(cd.Resources[1].ExtraIdentity)).To(Equal(0))
		})

		It("signs defaulted", func() {
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv)))
			cd := cv.GetDescriptor()
			Expect(len(cd.Resources)).To(Equal(2))
			Expect(len(cd.Resources[0].ExtraIdentity)).To(Equal(1))
			Expect(len(cd.Resources[1].ExtraIdentity)).To(Equal(1))
		})
	})
})
