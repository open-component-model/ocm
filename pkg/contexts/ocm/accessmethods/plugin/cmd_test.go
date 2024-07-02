package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/testutils"
	. "github.com/open-component-model/ocm/pkg/env"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
)

const (
	CA      = "/tmp/ca"
	VERSION = "v1"
)

var _ = Describe("Add with new access method", func() {
	var env *Environment
	var ctx ocm.Context
	var registry plugins.Set
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewEnvironment(TestData())
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

	It("handles resource options", func() {
		at := ctx.AccessMethods().GetType("test")
		Expect(at).NotTo(BeNil())

		h := at.ConfigOptionTypeSetHandler()
		Expect(h).NotTo(BeNil())
		Expect(h.GetName()).To(Equal("test"))

		ot := h.OptionTypes()
		Expect(len(ot)).To(Equal(2))

		opts := h.CreateOptions()
		Expect(sliceutils.Transform(opts.Options(), transformer.GetName[flagsets.Option, string])).To(ConsistOf(
			"mediaType", "accessPath"))

		fs := &pflag.FlagSet{}
		fs.SortFlags = true
		opts.AddFlags(fs)

		Expect("\n" + fs.FlagUsages()).To(Equal(`
      --accessPath string   file path
      --mediaType string    media type for artifact blob representation
`))

		MustBeSuccessful(fs.Parse([]string{"--accessPath", "filepath", "--" + options.MediatypeOption.GetName(), "yaml"}))

		cfg := flagsets.Config{}
		MustBeSuccessful(h.ApplyConfig(opts, cfg))
		Expect(cfg).To(YAMLEqual(`
mediaType: yaml
path: filepath
`))
	})
})
