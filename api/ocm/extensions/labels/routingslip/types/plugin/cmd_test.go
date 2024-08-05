package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/api/ocm/plugin/testutils"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/plugin/registration"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	PROVIDER = "acme.org"
)

var _ = Describe("Test Environment", func() {
	var env *Environment
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewEnvironment(TestData())

		ctx := env.OCMContext()
		plugins = Must(ConfigureTestPlugins(env, "testdata"))
		registry := plugincacheattr.Get(ctx)
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		plugins.Cleanup()
		env.Cleanup()
	})

	It("handles plugin based entry type", func() {
		prov := routingslip.For(env.OCMContext()).CreateConfigTypeSetConfigProvider()
		configopts := prov.CreateOptions()
		Expect(sliceutils.Transform(configopts.Options(), transformer.GetName[flagsets.Option, string])).To(ConsistOf(
			"entry", "comment", // default settings
			"mediaType", "accessPath", // by plugin
		))

		fs := &pflag.FlagSet{}
		fs.SortFlags = true
		configopts.AddFlags(fs)
		Expect("\n" + fs.FlagUsages()).To(Equal(`
      --accessPath string   file path
      --comment string      comment field value
      --entry YAML          routing slip entry specification (YAML)
      --mediaType string    media type for artifact blob representation
`))
		MustBeSuccessful(fs.Parse([]string{"--accessPath", "some path", "--" + options.MediatypeOption.GetName(), "media type"}))
		prov.SetTypeName("test")
		data := Must(prov.GetConfigFor(configopts))
		Expect(data).To(YAMLEqual(`
type: test
mediaType: media type
path: some path
`))
	})
})
