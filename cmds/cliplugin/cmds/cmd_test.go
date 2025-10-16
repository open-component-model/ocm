//go:build unix

package cmds_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/logging/logrusl"
	"github.com/mandelsoft/logging/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/version"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const KIND = "rhubarb"

var _ = Describe("cliplugin", func() {
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

		It("run plugin based ocm command", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("--config", "testdata/config.yaml", "check", KIND, "-d", "jul/10"))

			Expect("\n" + buf.String()).To(Equal(`
Yeah, it's rhabarb season - happy rhabarbing!
`))
		})

		It("runs plugin based ocm command with log", func() {
			var stdout bytes.Buffer
			var stdlog bytes.Buffer

			lctx := env.OCMContext().LoggingContext()
			lctx.SetBaseLogger(logrusl.WithWriter(utils.NewSyncWriter(&stdlog)).NewLogr())
			MustBeSuccessful(env.CatchOutput(&stdout).
				Execute("--config", "testdata/logcfg.yaml", "check", KIND, "-d", "jul/10"))

			Expect("\n" + stdout.String()).To(Equal(`
Yeah, it's rhabarb season - happy rhabarbing!
`))
			// {"date":".*","level":"debug","msg":"testing rhabarb season","realm":"cliplugin/rhabarber","time":".*"}
			Expect(stdlog.String()).To(StringMatchTrimmedWithContext(`
[^ ]* debug   \[cliplugin/rhabarber\] "testing rhabarb season" date="[^"]*"
`))
		})

		It("fails for undeclared config", func() {
			var buf bytes.Buffer

			Expect(env.CatchOutput(&buf).Execute("--config", "testdata/err.yaml", "check", KIND, "-d", "jul/10")).To(
				MatchError(`config type "err.config.acme.org" is unknown`))
		})

		It("shows command help", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("check", KIND, "--help"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm check rhubarb — Determine Whether We Are In Rhubarb Season

Synopsis:
  ocm check rhubarb <options>

Options:
  -d, --date string   the date to ask for (MM/DD)
  -h, --help          help for check

Description:
  The rhubarb season is between march and april.

`))
		})

		It("shows command help from main command", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("help", "check", KIND))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
ocm check rhubarb — Determine Whether We Are In Rhubarb Season

Synopsis:
  ocm check rhubarb <options>

Options:
  -d, --date string   the date to ask for (MM/DD)
  -h, --help          help for check

Description:
  The rhubarb season is between march and april.

`))
		})

		It("describe", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("describe", "plugin", "cliplugin"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Plugin Name:      cliplugin
Plugin Version:   ` + version.Get().String() + `
Path:             ` + plugins.Path() + `/cliplugin
Status:           valid
Source:           manually installed
Capabilities:     CLI Commands, Config Types
Description: 
      The plugin offers the check command for object type rhubarb to check the rhubarb season.

CLI Extensions:
- Name:   check (determine whether we are in rhubarb season)
  Object: rhubarb
  Verb:   check
  Usage:  check rhubarb <options>
    The rhubarb season is between march and april.

Config Types for CLI Command Extensions:
- Name: rhabarber.config.acme.org
    The config type «rhabarber.config.acme.org» can be used to configure the season for rhubarb:
    
        type: rhabarber.config.acme.org
        start: mar/1
        end: apr/30

  Versions:
  - Version: v1
*** found 1 plugins
`))
		})
	})
})
