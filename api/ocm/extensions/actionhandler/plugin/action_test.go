//go:build unix

package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext/action/handlers"
	. "ocm.software/ocm/api/helper/builder"
	oci_repository_prepare "ocm.software/ocm/api/oci/extensions/actions/oci-repository-prepare"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/actionhandler/plugin"
	"ocm.software/ocm/api/ocm/plugin/plugins"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
)

const PLUGIN = "test"

var _ = Describe("plugin action handler", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Builder
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))
		p := registry.Get("action")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("executes with no plugin registration", func() {
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "ghcr.io", "mandelsoft", nil))
		Expect(result).To(BeNil())
	})

	It("executes with no handler", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "mandelsoft.org", "mandelsoft", nil))
		Expect(result).To(BeNil())
	})

	It("used default registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction("test"), handlers.ForSelectors("mandelsoft.org"))
		MustFailWithMessage(plugin.RegisterActionHandler(ctx.AttributesContext(), "action", ctx, opts), "action \"test\" is unknown for plugin action")
	})

	It("uses default registration", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "ghcr.io", "mandelsoft", nil))
		Expect(result).NotTo(BeNil())
		Expect(result.Message).To(Equal("all good"))
	})

	It("uses default pattern registration", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "xyz.dkr.ecr.us-west-2.amazonaws.com", "mandelsoft", nil))
		Expect(result).NotTo(BeNil())
		Expect(result.Message).To(Equal("all good"))
	})

	It("executes action for dynamic registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction(oci_repository_prepare.Type), handlers.ForSelectors("mandelsoft.org"))
		MustBeSuccessful(plugin.RegisterActionHandler(ctx.AttributesContext(), "action", ctx, opts))

		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "mandelsoft.org", "mandelsoft", nil))
		Expect(result.Message).To(Equal("all good"))
	})

	It("executed action after abstract registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction(oci_repository_prepare.Type), handlers.ForSelectors("mandelsoft.org"))
		ok := Must(ctx.GetActions().RegisterByName("plugin/action", ctx.OCIContext(), ctx, opts))
		Expect(ok).To(BeTrue())
	})
})
