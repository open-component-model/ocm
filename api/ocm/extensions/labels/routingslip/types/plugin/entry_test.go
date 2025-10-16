//go:build unix

package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	. "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/spi"
	"ocm.software/ocm/api/ocm/plugin/plugins"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Environment
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewEnvironment()
		ctx = env.OCMContext()
		plugins, registry = Must2(ConfigureTestPlugins2(env, "testdata"))
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("registers valuesets", func() {
		scheme := spi.For(env.OCMContext())
		Expect(scheme.GetType("test")).NotTo(BeNil())
	})

	It("validates valuesets", func() {
		scheme := spi.For(env.OCMContext())
		t := scheme.GetType("test")
		Expect(t).NotTo(BeNil())
		opts := t.ConfigOptionTypeSetHandler().CreateOptions()
		var fs pflag.FlagSet
		opts.AddFlags(&fs)

		NotNil(fs.Lookup("accessPath"))
		MustBeSuccessful(fs.Set("accessPath", "somepath"))
		NotNil(fs.Lookup("mediaType"))
		MustBeSuccessful(fs.Set("mediaType", "a simple test"))

		cfg := flagsets.Config{}
		MustBeSuccessful(t.ConfigOptionTypeSetHandler().ApplyConfig(opts, cfg))

		Expect(cfg).To(Equal(flagsets.Config{
			"path":      "somepath",
			"mediaType": "a simple test",
		}))
	})
})
