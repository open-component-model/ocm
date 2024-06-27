//go:build unix

package cmds_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
)

var _ = Describe("subcmdplugin", func() {
	Context("lib", func() {
		var env *TestEnv
		var plugins TempPluginDir

		BeforeEach(func() {
			env = NewTestEnv(TestData())
			plugins = Must(ConfigureTestPlugins(env, "testdata/plugins"))

			registry := plugincacheattr.Get(env)
			//	Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("cliplugin")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		AfterEach(func() {
			plugins.Cleanup()
			env.Cleanup()
		})

		Context("local help", func() {
			It("shows group command help", func() {
				var buf bytes.Buffer

				MustBeSuccessful(env.CatchOutput(&buf).Execute("group", "--help"))
				Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm group — A Provided Command Group

Synopsis:
  ocm group <options>

Available Commands:
  demo        a demo command

Flags:
  -h, --help   help for group

Description:
  A provided command group with a demo command
  Use ocm group <command> -h for additional help.
`))
			})

			It("shows sub command help", func() {
				var buf bytes.Buffer

				MustBeSuccessful(env.CatchOutput(&buf).Execute("group", "demo", "--help"))
				Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm group demo — A Demo Command

Synopsis:
  ocm group demo <options> [flags]

Flags:
  -h, --help             help for demo
      --version string   some overloaded option

Description:
  a demo command in a provided command group
`))
			})
		})

		Context("main help", func() {
			It("shows group command help", func() {
				var buf bytes.Buffer

				MustBeSuccessful(env.CatchOutput(&buf).Execute("help", "group"))
				Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm group — A Provided Command Group

Synopsis:
  ocm group <options>

Available Commands:
  demo        a demo command

Flags:
  -h, --help   help for group

Description:
  A provided command group with a demo command
  Use ocm group <command> -h for additional help.
`))
			})

			It("shows sub command help", func() {
				var buf bytes.Buffer

				MustBeSuccessful(env.CatchOutput(&buf).Execute("help", "group", "demo"))
				Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm group demo — A Demo Command

Synopsis:
  ocm group demo <options> [flags]

Flags:
  -h, --help             help for demo
      --version string   some overloaded option

Description:
  a demo command in a provided command group
`))
			})
		})
	})
})
