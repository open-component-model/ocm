//go:build unix

package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const PLUGINS = "/testdata"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var plugins TempPluginDir

	BeforeEach(func() {
		env = NewTestEnv()
		plugins = Must(ConfigureTestPlugins(env, "testdata"))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get plugins", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+plugins.Path(), "get", "plugins")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
PLUGIN VERSION SOURCE DESCRIPTION                    CAPABILITIES
test   v1      local  a test plugin without function Access Methods
`))
	})
	It("get plugins with additional info", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+plugins.Path(), "get", "plugins", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
PLUGIN VERSION SOURCE DESCRIPTION                    ACCESSMETHODS UPLOADERS DOWNLOADERS ACTIONS
test   v1      local  a test plugin without function test[v1]
`))
	})
})
