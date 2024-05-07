package genericocireg_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/finalizer"
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
