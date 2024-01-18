// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package identity_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/helm/identity"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	ociidentity "github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/oci/identity"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
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
