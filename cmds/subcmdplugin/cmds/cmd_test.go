//go:build unix

package cmds_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"
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

		Context("error handling", func() {
			It("provides error", func() {
				var buf bytes.Buffer

				ExpectError(env.CatchOutput(&buf).Execute("group", "demo", "--version=error")).To(MatchError("error processing plugin command command: demo error"))
			})

			It("provides error and error outputp", func() {
				var buf bytes.Buffer

				ExpectError(env.CatchOutput(&buf).Execute("group", "demo", "--version=error this is an error my friend")).To(MatchError(`error processing plugin command command: demo error: with stderr
this is an error my friend`))
			})
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

Options:
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

Options:
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

Options:
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

Options:
  -h, --help             help for demo
      --version string   some overloaded option

Description:
  a demo command in a provided command group
`))
			})
		})
	})
})
