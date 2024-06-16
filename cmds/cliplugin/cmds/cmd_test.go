//go:build unix

package cmds_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/logging/logrusl"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/version"
)

const (
	ARCH = "/tmp/ctf"
	COMP = "acme.org/test"
	VERS = "1.0.0"
	PROV = "acme.org"
)

var _ = Describe("cliplugin", func() {
	Context("lib", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(TestData())

			cache.DirectoryCache.Reset()
			plugindirattr.Set(env.OCMContext(), "testdata/plugins")

			registry := plugincacheattr.Get(env)
			//	Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("cliplugin")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("run plugin based ocm command", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("--config", "testdata/config.yaml", "rhabarber", "-d", "jul/10"))

			Expect("\n" + buf.String()).To(Equal(`
Yeah, it's rhabarb season - happy rhabarbing!
`))
		})

		It("runs plugin based ocm command with log", func() {
			var stdout bytes.Buffer
			var stdlog bytes.Buffer

			lctx := env.OCMContext().LoggingContext()
			lctx.SetBaseLogger(logrusl.WithWriter(&stdlog).NewLogr())
			MustBeSuccessful(env.CatchOutput(&stdout).
				Execute("--config", "testdata/logcfg.yaml", "rhabarber", "-d", "jul/10"))

			Expect("\n" + stdout.String()).To(Equal(`
Yeah, it's rhabarb season - happy rhabarbing!
`))
			// {"date":".*","level":"debug","msg":"testing rhabarb season","realm":"cliplugin/rhabarber","time":".*"}
			Expect(stdlog.String()).To(StringMatchTrimmedWithContext(`
.{25} debug   \[cliplugin/rhabarber\] "testing rhabarb season" date=".{30}"
`))
		})

		It("fails for undeclared confiug", func() {
			var buf bytes.Buffer

			Expect(env.CatchOutput(&buf).Execute("--config", "testdata/err.yaml", "rhabarber", "-d", "jul/10")).To(
				MatchError(`config type "err.config.acme.org" is unknown`))
		})

		It("describe", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("describe", "plugin", "cliplugin"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Plugin Name:      cliplugin
Plugin Version:   ` + version.Get().String() + `
Path:             testdata/plugins/cliplugin
Status:           valid
Source:           manually installed
Capabilities:     CLI Commands, Config Types
Description: 
      The plugin offers the top-level command rhabarber

CLI Extensions:
- Name: rhabarber (determine whether we are in rhubarb season)
  Usage: rhabarber <options>
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
