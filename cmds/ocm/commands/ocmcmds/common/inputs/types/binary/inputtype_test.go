package binary

import (
	. "github.com/onsi/ginkgo/v2"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	. "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"
)

var _ = Describe("Input Type", func() {
	var env *InputTest

	BeforeEach(func() {
		env = NewInputTest(TYPE)
	})

	It("simple string decode", func() {
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.DataOption, "!stringdata")
		env.Check(&Spec{
			Data:        runtime.Binary("stringdata"),
			ProcessSpec: cpi.NewProcessSpec("media", true),
		})
	})

	It("binary decode", func() {
		env.Set(options.CompressOption, "true")
		env.Set(options.MediaTypeOption, "media")
		env.Set(options.DataOption, "IXN0cmluZ2RhdGE=")
		env.Check(&Spec{
			Data:        runtime.Binary("!stringdata"),
			ProcessSpec: cpi.NewProcessSpec("media", true),
		})
	})
})
