package signing_test

import (
	. "github.com/onsi/ginkgo/v2"

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
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Simple signing handlers", func() {
	Context("", func() {
		ctx := ocm.DefaultContext()

		var cv ocm.ComponentVersionAccess
		var pub signutils.GenericPublicKey
		var priv signutils.GenericPrivateKey

		BeforeEach(func() {
			priv, pub = Must2(rsa.CreateKeyPair())
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
})
