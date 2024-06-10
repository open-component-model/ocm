//go:build unix

package plugin_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/spi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
)

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Environment

	BeforeEach(func() {
		env = NewEnvironment()
		ctx = env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
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
