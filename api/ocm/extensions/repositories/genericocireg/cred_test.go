package genericocireg_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/credentials"
	ociidentity "ocm.software/ocm/api/credentials/builtin/oci/identity"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	common "ocm.software/ocm/api/utils/misc"
)

var _ = Describe("consumer id handling", func() {
	It("creates a dummy component", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		ctx := ocm.New(datacontext.MODE_EXTENDED)
		credctx := ctx.CredentialsContext()

		creds := ociidentity.SimpleCredentials("test", "password")
		spec := ocireg.NewRepositorySpec("ghcr.io/open-component-model/test")

		id := credentials.GetProvidedConsumerId(spec, credentials.StringUsageContext(COMPONENT))
		Expect(id).To(Equal(credentials.NewConsumerIdentity(ociidentity.CONSUMER_TYPE, ociidentity.ID_HOSTNAME, "ghcr.io", ociidentity.ID_PATHPREFIX, "open-component-model/test/component-descriptors/"+COMPONENT)))

		id = credentials.GetProvidedConsumerId(spec)
		Expect(id).To(Equal(credentials.NewConsumerIdentity(ociidentity.CONSUMER_TYPE, ociidentity.ID_HOSTNAME, "ghcr.io", ociidentity.ID_PATHPREFIX, "open-component-model/test")))

		credctx.SetCredentialsForConsumer(id, creds)

		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))

		id = credentials.GetProvidedConsumerId(repo, credentials.StringUsageContext(COMPONENT))
		Expect(id).To(Equal(credentials.NewConsumerIdentity(ociidentity.CONSUMER_TYPE, ociidentity.ID_HOSTNAME, "ghcr.io", ociidentity.ID_PATHPREFIX, "open-component-model/test/component-descriptors/"+COMPONENT)))

		creds = Must(credentials.CredentialsForConsumer(credctx, id))

		Expect(creds.Properties()).To(Equal(common.Properties{
			ociidentity.ATTR_USERNAME: "test",
			ociidentity.ATTR_PASSWORD: "password",
		}))
	})
})
