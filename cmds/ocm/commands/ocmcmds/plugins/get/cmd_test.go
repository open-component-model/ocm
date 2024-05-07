//go:build unix

package get_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/filepath/pkg/filepath"
)

const PLUGINS = "/testdata"

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var path string

	BeforeEach(func() {
		env = NewTestEnv(TestData())

		// use os filesystem here
		p, err := filepath.Abs("testdata")
		Expect(err).To(Succeed())
		path = p
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get plugins", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+path, "get", "plugins")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
PLUGIN VERSION SOURCE DESCRIPTION                    CAPABILITIES
test   v1      local  a test plugin without function accessmethods
`))
	})
	It("get plugins with additional info", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+path, "get", "plugins", "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
PLUGIN VERSION SOURCE DESCRIPTION                    ACCESSMETHODS UPLOADERS DOWNLOADERS ACTIONS
test   v1      local  a test plugin without function test[v1]
`))
	})

})
