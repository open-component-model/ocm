//go:build unix

package valuesets_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/api/ocm/plugin/testutils"

	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/repositories/composition"
	"ocm.software/ocm/api/ocm/plugin/registration"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
)

const (
	ARCH = "/tmp/ctf"
	COMP = "acme.org/test"
	VERS = "1.0.0"
	PROV = "acme.org"
)

var _ = Describe("demoplugin", func() {
	Context("lib", func() {
		var env *Builder
		var plugins TempPluginDir

		BeforeEach(func() {
			env = NewBuilder(TestData())
			plugins = Must(ConfigureTestPlugins(env, "testdata"))

			registry := plugincacheattr.Get(env)
			Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("demo")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))

			env.OCMCompositionRepository("test", func() {
				env.ComponentVersion(COMP, VERS, func() {
					env.Provider(PROV)
				})
			})
			env.RSAKeyPair(PROV)
		})

		AfterEach(func() {
			plugins.Cleanup()
			env.Cleanup()
		})

		It("add check routing slip entry", func() {
			fs := &pflag.FlagSet{}
			prov := routingslip.For(env).CreateConfigTypeSetConfigProvider()
			configopts := prov.CreateOptions()
			configopts.AddFlags(fs)

			MustBeSuccessful(fs.Parse([]string{"--checkStatus", "test=passed", "--checkMessage", "test=25 tests successful"}))
			prov.SetTypeName("check")
			data := Must(prov.GetConfigFor(configopts))

			Expect(data).To(YAMLEqual(`
type: check
checks:
  test:
    status: passed
    message: 25 tests successful
`))

			entry := Must(routingslip.NewGenericEntry("", data))
			MustBeSuccessful(entry.Validate(env.OCMContext()))

			repo := composition.NewRepository(env, "test")
			defer Close(repo, "repo")

			cv := Must(repo.LookupComponentVersion(COMP, VERS))
			defer Close(cv, "cv")

			Must(routingslip.AddEntry(cv, PROV, rsa.Algorithm, entry, nil))
		})
	})
})
