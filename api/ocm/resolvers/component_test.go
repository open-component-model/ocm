package resolvers_test

import (
	"github.com/mandelsoft/goutils/sliceutils"
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	ARCH1      = "ctf1"
	ARCH2      = "ctf2"
	COMPONENT1 = "acme.org/test1"
	COMPONENT2 = "acme.org/test2"
	VERSION1   = "1.0.0"
	VERSION2   = "1.1.0"
)

var _ = Describe("resolver", func() {
	var env *Builder

	var spec1 ocm.RepositorySpec
	var spec2 ocm.RepositorySpec

	BeforeEach(func() {
		env = NewBuilder()

		env.OCMCommonTransport(ARCH1, accessio.FormatDirectory, func() {
			env.Component(COMPONENT1, func() {
				env.Version(VERSION1, func() {
				})
			})
			env.Component(COMPONENT2, func() {
				env.Version(VERSION1, func() {
				})
			})
		})
		env.OCMCommonTransport(ARCH2, accessio.FormatDirectory, func() {
			env.Component(COMPONENT1, func() {
				env.Version(VERSION2, func() {
				})
			})
		})

		spec1 = Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH1, env))
		spec2 = Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH2, env))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lookup in multiple providers", func() {
		ctx := ocm.New()

		repo1 := Must(ctx.RepositoryForSpec(spec1))
		defer Close(repo1, "repo1")
		repo2 := Must(ctx.RepositoryForSpec(spec2))
		defer Close(repo2, "repo2")
		resolver := resolvers.NewCompoundComponentResolver(
			resolvers.ComponentResolverForRepository(repo1),
			resolvers.ComponentResolverForRepository(repo2),
		)

		list := Must(resolvers.ListComponentVersions(COMPONENT1, resolver))
		Expect(list).To(Equal(sliceutils.AsSlice("1.0.0", "1.1.0")))

		list = Must(resolvers.ListComponentVersions(COMPONENT2, resolver))
		Expect(list).To(Equal(sliceutils.AsSlice("1.0.0")))
	})

	It("lookup cv in second", func() {
		ctx := ocm.New()

		repo1 := Must(ctx.RepositoryForSpec(spec1))
		defer Close(repo1, "repo1")
		repo2 := Must(ctx.RepositoryForSpec(spec2))
		defer Close(repo2, "repo2")
		resolver := resolvers.ComponentVersionResolverForComponentResolver(resolvers.NewCompoundComponentResolver(
			resolvers.ComponentResolverForRepository(repo1),
			resolvers.ComponentResolverForRepository(repo2),
		))

		cv := Must(resolver.LookupComponentVersion(COMPONENT1, VERSION2))
		defer Close(cv)
		Expect(cv).NotTo(BeNil())
	})
})
