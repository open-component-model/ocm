//go:build unix

package describe_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/testutils"
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
		plugins.Cleanup()
		env.Cleanup()
	})

	It("get plugins", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("-X", "plugindir="+plugins.Path(), "describe", "plugins")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
Plugin Name:      action
Plugin Version:   v1
Path:             ` + plugins.Path() + `/action
Status:           valid
Source:           manually installed
Capabilities:     Actions
Description: 
      a test plugin with action oci.repository.prepare

Actions:
- Name: oci.repository.prepare
    Prepare the usage of a repository in an OCI registry.

    The hostname of the target repository is used as selector. The action should
    assure, that the requested repository is available on the target OCI registry.
    
    Spec version v1 uses the following specification fields:
    - «hostname» *string*: The  hostname of the OCI registry.
    - «repository» *string*: The OCI repository name.
  Info:
    test action
  Versions:
  - v1 (best matching)
  Handler accepts standard credentials
----------------------
Plugin Name:      test
Plugin Version:   v1
Path:             ` + plugins.Path() + `/test
Status:           valid
Source:           manually installed
Capabilities:     Access Methods
Description: 
      a test plugin with access method test

Access Methods:
- Name: test
  Versions:
  - Version: v1
*** found 2 plugins
`))
	})
})
