package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	. "ocm.software/ocm/api/tech/helm/identity"
	ociidentity "ocm.software/ocm/api/tech/oci/identity"
	common "ocm.software/ocm/api/utils/misc"
)

var _ = Describe("consumer id handling", func() {
	Context("id deternation", func() {
		It("handles helm repos", func() {
			id := GetConsumerId("https://acme.org/charts", "demo:v1")
			Expect(id).To(Equal(credentials.NewConsumerIdentity(CONSUMER_TYPE,
				"pathprefix", "charts",
				"port", "443",
				"hostname", "acme.org",
				"scheme", "https",
			)))
		})

		It("handles oci repos", func() {
			id := GetConsumerId("oci://acme.org/charts", "demo:v1")
			Expect(id).To(Equal(credentials.NewConsumerIdentity(ociidentity.CONSUMER_TYPE,
				"pathprefix", "charts/demo",
				"hostname", "acme.org",
			)))
		})
	})

	Context("query credentials", func() {
		var ctx oci.Context
		var credctx credentials.Context

		BeforeEach(func() {
			ctx = oci.New(datacontext.MODE_EXTENDED)
			credctx = ctx.CredentialsContext()
		})

		It("queries helm credentials", func() {
			id := GetConsumerId("https://acme.org/charts", "demo:v1")
			credctx.SetCredentialsForConsumer(id,
				credentials.CredentialsFromList(
					ATTR_USERNAME, "helm",
					ATTR_PASSWORD, "helmpass",
				),
			)

			creds := GetCredentials(ctx, "https://acme.org/charts", "demo:v1")
			Expect(creds).To(Equal(common.Properties{
				ATTR_USERNAME: "helm",
				ATTR_PASSWORD: "helmpass",
			}))
		})

		It("queries oci credentials", func() {
			id := GetConsumerId("oci://acme.org/charts", "demo:v1")
			credctx.SetCredentialsForConsumer(id,
				credentials.CredentialsFromList(
					ATTR_USERNAME, "oci",
					ATTR_PASSWORD, "ocipass",
				),
			)

			creds := GetCredentials(ctx, "oci://acme.org/charts", "demo:v1")
			Expect(creds).To(Equal(common.Properties{
				ATTR_USERNAME: "oci",
				ATTR_PASSWORD: "ocipass",
			}))
		})
	})
})
