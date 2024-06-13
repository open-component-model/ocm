//go:build unix

package cmds_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/version"
)

const ARCH = "/tmp/ctf"
const COMP = "acme.org/test"
const VERS = "1.0.0"
const PROV = "acme.org"

var _ = Describe("cliplugin", func() {

	Context("lib", func() {
		var env *TestEnv

		BeforeEach(func() {
			env = NewTestEnv(TestData())

			cache.DirectoryCache.Reset()
			plugindirattr.Set(env.OCMContext(), "testdata")

			registry := plugincacheattr.Get(env)
			Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get("cliplugin")
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("run plugin based ocm command", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("rhabarber", "-d", "apr/1"))

			Expect("\n" + buf.String()).To(Equal(`
Yeah, it's rhabarb season - happy rhabarbing!
`))
		})

		It("describe", func() {
			var buf bytes.Buffer

			MustBeSuccessful(env.CatchOutput(&buf).Execute("describe", "plugin", "cliplugin"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
Plugin Name:      cliplugin
Plugin Version:   ` + version.Get().String() + `
Path:             testdata/cliplugin
Status:           valid
Source:           manually installed
Capabilities:     CLI Commands
Description: 
      The plugin offers the top-level command rhabarber

CLI Extensions:
- Name: rhabarber (determine whether we are in rhubarb season)
  Usage: rhabarber <options>
    The rhubarb season is between march and april.
*** found 1 plugins
`))
		})
	})
})
