//go:build unix

package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/ocm/valuemergehandler"
	"ocm.software/ocm/api/ocm/valuemergehandler/handlers/defaultmerge"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
)

const (
	PLUGIN    = "merge"
	ALGORITHM = "acme.org/test"
)

var _ = Describe("plugin value merge handler", func() {
	var ctx ocm.Context
	var env *Builder
	var plugins TempPluginDir
	var registry valuemergehandler.Registry

	BeforeEach(func() {
		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugins = Must(ConfigureTestPlugins(ctx, "testdata"))
		registry = valuemergehandler.For(ctx)
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("executes handler", func() {
		registration.RegisterExtensions(ctx)

		Expect(registry.GetHandler(ALGORITHM)).NotTo(BeNil())

		spec := Must(valuemergehandler.NewSpecification(ALGORITHM, defaultmerge.NewConfig("test")))
		var local, inbound valuemergehandler.Value

		local.SetValue("local")
		inbound.SetValue("inbound")
		mod := Must(valuemergehandler.Merge(ctx, spec, "", local, &inbound))

		Expect(mod).To(BeTrue())
		Expect(inbound.RawMessage).To(YAMLEqual(`{"mode":"resolved"}`))
	})

	It("assigns specs", func() {
		registration.RegisterExtensions(ctx)

		Expect(registry.GetHandler(ALGORITHM)).NotTo(BeNil())

		s := registry.GetAssignment(hpi.LabelHint("testlabel", "v2"))
		Expect(s).NotTo(BeNil())
		Expect(s.Algorithm).To(Equal("simpleMapMerge"))
	})
})
