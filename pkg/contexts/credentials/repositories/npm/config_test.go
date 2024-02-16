package npm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/builtin/npm/identity"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/npm"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("Config deserialization Test Environment", func() {
	It("read .npmrc", func() {
		ctx := credentials.New()

		repo := Must(npm.NewRepository(ctx, "testdata/.npmrc"))
		Expect(Must(repo.LookupCredentials("https://registry.npmjs.org")).Properties()).To(Equal(common.Properties{identity.ATTR_TOKEN: "npm_TOKEN"}))
		Expect(Must(repo.LookupCredentials("https://npm.registry.acme.com/api/npm")).Properties()).To(Equal(common.Properties{identity.ATTR_TOKEN: "bearer_TOKEN"}))
	})

	It("propagates credentials", func() {
		ctx := credentials.New()

		spec := npm.NewRepositorySpec("testdata/.npmrc")

		_ = Must(ctx.RepositoryForSpec(spec))
		id := identity.GetConsumerId("https://registry.npmjs.org", "pkg")

		creds := Must(credentials.CredentialsForConsumer(ctx, id))
		Expect(creds).NotTo(BeNil())
		Expect(creds.GetProperty(identity.ATTR_TOKEN)).To(Equal("npm_TOKEN"))
	})

	It("has description", func() {
		ctx := credentials.New()
		t := ctx.RepositoryTypes().GetType(npm.TypeV1)
		Expect(t).NotTo(BeNil())
		Expect(t.Description()).NotTo(Equal(""))
	})

})
