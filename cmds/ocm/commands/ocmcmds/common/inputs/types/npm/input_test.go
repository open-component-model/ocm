package npm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"

	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/npm"
)

var _ = Describe("Input Type", func() {
	var env *InputTest

	BeforeEach(func() {
		env = NewInputTest(npm.TYPE)
	})

	It("simple fetch", func() {
		env.Set(options.RegistryOption, "https://registry.npmjs.org")
		env.Set(options.PackageOption, "yargs")
		env.Set(options.PackageVersionOption, "17.7.1")
		env.Check(&npm.Spec{
			InputSpecBase: inputs.InputSpecBase{},
			Registry:      "https://registry.npmjs.org",
			Package:       "yargs",
			Version:       "17.7.1",
		})
	})
})
