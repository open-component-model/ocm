//go:build unix

package valuesets_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const ARCH = "/tmp/ctf"
const COMP = "acme.org/test"
const VERS = "1.0.0"
const PROV = "acme.org"

var _ = Describe("demoplugin", func() {
	/*
			Context("cli", func() {
				var env *testhelper.TestEnv

				BeforeEach(func() {
					env = testhelper.NewTestEnv(testhelper.TestData())

					cache.DirectoryCache.Reset()
					plugindirattr.Set(env.OCMContext(), "testdata")

					env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
						env.ComponentVersion(COMP, VERS, func() {
							env.Provider(PROV)
						})
					})
					env.RSAKeyPair(PROV)
				})

				AfterEach(func() {
					env.Cleanup()
				})

				It("add check routing slip entry", func() {
					buf := bytes.NewBuffer(nil)
					MustBeSuccessful(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROV, "check", "--checkStatus", "test=passed", "--checkMessage", "test=25 tests successful"))
					Expect(buf.String()).To(Equal(""))

					buf.Reset()
					MustBeSuccessful(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, PROV))
					Expect(buf.String()).To(StringMatchTrimmedWithContext(`
		COMPONENT-VERSION   NAME     TYPE  TIMESTAMP            DESCRIPTION
		acme.org/test:1.0.0 acme.org check .{20} test: passed
		`))
					buf.Reset()
					MustBeSuccessful(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, PROV, "-oyaml"))
					Expect(buf.String()).To(StringMatchTrimmedWithContext(`message: 25 tests successful`))
				})
			})
	*/

	Context("lib", func() {
		var env *Builder

		BeforeEach(func() {
			env = NewBuilder(TestData())

			cache.DirectoryCache.Reset()
			plugindirattr.Set(env.OCMContext(), "testdata")

			registry := plugincacheattr.Get(env)
			Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("demo")
			Expect(p).NotTo(BeNil())

			env.OCMCompositionRepository("test", func() {
				env.ComponentVersion(COMP, VERS, func() {
					env.Provider(PROV)
				})
			})
			env.RSAKeyPair(PROV)
		})

		AfterEach(func() {
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
