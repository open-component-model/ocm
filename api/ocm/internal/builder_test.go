package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/config"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	local "ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/utils/runtime"
)

var _ = Describe("builder test", func() {
	It("creates local", func() {
		ctx := local.Builder{}.New(datacontext.MODE_SHARED)

		Expect(ctx.AttributesContext()).To(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(BaseRepoTypes(ctx.RepositoryTypes())).To(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(ctx.AccessMethods()).To(BeIdenticalTo(local.DefaultAccessTypeScheme))
		Expect(ctx.RepositorySpecHandlers()).To(BeIdenticalTo(local.DefaultRepositorySpecHandlers))
		Expect(ctx.BlobHandlers()).To(BeIdenticalTo(local.DefaultBlobHandlerRegistry))
		Expect(ctx.BlobDigesters()).To(BeIdenticalTo(local.DefaultBlobDigesterRegistry))

		Expect(ctx.ConfigContext().GetId()).To(BeIdenticalTo(config.DefaultContext().GetId()))
		Expect(ctx.CredentialsContext().GetId()).To(BeIdenticalTo(credentials.DefaultContext().GetId()))
		Expect(ctx.OCIContext().GetId()).To(BeIdenticalTo(oci.DefaultContext().GetId()))
	})

	It("creates defaulted", func() {
		ctx := local.Builder{}.New(datacontext.MODE_DEFAULTED)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(BaseRepoTypes(ctx.RepositoryTypes())).To(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(ctx.AccessMethods()).To(BeIdenticalTo(local.DefaultAccessTypeScheme))
		Expect(ctx.RepositorySpecHandlers()).To(BeIdenticalTo(local.DefaultRepositorySpecHandlers))
		Expect(ctx.BlobHandlers()).To(BeIdenticalTo(local.DefaultBlobHandlerRegistry))
		Expect(ctx.BlobDigesters()).To(BeIdenticalTo(local.DefaultBlobDigesterRegistry))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.ConfigContext().ConfigTypes()).To(BeIdenticalTo(config.DefaultContext().ConfigTypes()))

		Expect(ctx.CredentialsContext()).NotTo(BeIdenticalTo(credentials.DefaultContext()))
		Expect(ctx.CredentialsContext().RepositoryTypes()).To(BeIdenticalTo(credentials.DefaultContext().RepositoryTypes()))

		Expect(ctx.OCIContext()).NotTo(BeIdenticalTo(oci.DefaultContext()))
		Expect(ctx.OCIContext().RepositoryTypes()).To(BeIdenticalTo(oci.DefaultContext().RepositoryTypes()))
	})

	It("creates configured", func() {
		ctx := local.Builder{}.New(datacontext.MODE_CONFIGURED)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).NotTo(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(ctx.RepositoryTypes().KnownTypeNames()).To(Equal(local.DefaultRepositoryTypeScheme.KnownTypeNames()))
		Expect(ctx.AccessMethods()).NotTo(BeIdenticalTo(local.DefaultAccessTypeScheme))
		Expect(ctx.RepositorySpecHandlers()).NotTo(BeIdenticalTo(local.DefaultRepositorySpecHandlers))
		Expect(ctx.BlobHandlers()).NotTo(BeIdenticalTo(local.DefaultBlobHandlerRegistry))
		Expect(ctx.BlobDigesters()).NotTo(BeIdenticalTo(local.DefaultBlobDigesterRegistry))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.ConfigContext().ConfigTypes()).NotTo(BeIdenticalTo(config.DefaultContext().ConfigTypes()))
		Expect(ctx.ConfigContext().ConfigTypes().KnownTypeNames()).To(Equal(config.DefaultContext().ConfigTypes().KnownTypeNames()))

		Expect(ctx.CredentialsContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.CredentialsContext().RepositoryTypes()).NotTo(BeIdenticalTo(credentials.DefaultContext().RepositoryTypes()))
		Expect(ctx.CredentialsContext().RepositoryTypes().KnownTypeNames()).To(Equal(credentials.DefaultContext().RepositoryTypes().KnownTypeNames()))

		Expect(ctx.OCIContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.OCIContext().RepositoryTypes()).NotTo(BeIdenticalTo(oci.DefaultContext().RepositoryTypes()))
		Expect(ctx.OCIContext().RepositoryTypes().KnownTypeNames()).To(Equal(oci.DefaultContext().RepositoryTypes().KnownTypeNames()))
	})

	It("creates iniial", func() {
		ctx := local.Builder{}.New(datacontext.MODE_INITIAL)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).NotTo(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(len(ctx.RepositoryTypes().KnownTypeNames())).To(Equal(0))
		Expect(len(ctx.AccessMethods().KnownTypeNames())).To(Equal(0))
		Expect(len(ctx.RepositorySpecHandlers().KnownTypeNames())).To(Equal(0))
		Expect(ctx.BlobHandlers().IsInitial()).To(Equal(true))
		Expect(ctx.BlobDigesters().IsInitial()).To(Equal(true))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(len(ctx.ConfigContext().ConfigTypes().KnownTypeNames())).To(Equal(0))

		Expect(ctx.CredentialsContext()).NotTo(BeIdenticalTo(credentials.DefaultContext()))
		Expect(len(ctx.CredentialsContext().RepositoryTypes().KnownTypeNames())).To(Equal(0))
	})
})

func BaseRepoTypes(r cpi.RepositoryTypeScheme) runtime.Scheme[local.RepositorySpec, local.RepositoryType] {
	return r.BaseScheme()
}
