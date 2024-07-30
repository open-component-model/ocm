package signing_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	rsa_pss "github.com/open-component-model/ocm/pkg/signing/handlers/rsa-pss"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

var _ = Describe("Simple signing handlers", func() {
	ctx := ocm.DefaultContext()

	var cv ocm.ComponentVersionAccess
	var pub signutils.GenericPublicKey
	var priv signutils.GenericPrivateKey

	BeforeEach(func() {
		priv, pub = Must2(rsa.CreateKeyPair())
	})

	Context("", func() {
		BeforeEach(func() {
			cv = composition.NewComponentVersion(ctx, COMPONENTA, VERSION)
			MustBeSuccessful(cv.SetResourceBlob(ocm.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation), blobaccess.ForString(mime.MIME_TEXT, "test data"), "", nil))
		})

		DescribeTable("rsa handlers", func(kind string) {
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv)))
			Must(signing.VerifyComponentVersion(cv, "signature", signing.PublicKey("signature", pub)))
		},
			Entry("rsa", rsa.Algorithm),
			Entry("rsapss", rsa_pss.Algorithm),
		)
	})

	Context("non-unique resources", func() {
		BeforeEach(func() {
			cv = composition.NewComponentVersion(ctx, COMPONENTA, VERSION)

			meta := ocm.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation)
			meta.Version = "v1"
			meta.ExtraIdentity = map[string]string{}
			MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, "test data"), "", nil))
			meta.Version = "v2"
			MustBeSuccessful(cv.SetResourceBlob(meta, blobaccess.ForString(mime.MIME_TEXT, "other test data"), "", nil, ocm.TargetIndex(-1)))
		})

		It("signs without modification", func() {
			Must(signing.SignComponentVersion(cv, "signature", signing.PrivateKey("signature", priv)))
			cd := cv.GetDescriptor()
			Expect(len(cd.Resources)).To(Equal(2))
			Expect(len(cd.Resources[0].ExtraIdentity)).To(Equal(0))
			Expect(len(cd.Resources[1].ExtraIdentity)).To(Equal(0))
		})
	})
})
