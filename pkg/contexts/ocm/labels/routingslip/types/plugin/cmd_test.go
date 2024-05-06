package plugin_test

import (
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/options"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const PROVIDER = "acme.org"

var _ = Describe("Test Environment", func() {
	var env *Environment

	BeforeEach(func() {
		env = NewEnvironment(TestData())

		ctx := env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry := plugincacheattr.Get(ctx)
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
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
