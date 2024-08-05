package file

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "file"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		Usage("The path must denote a file relative the resources file. "), ConfigHandler()))
}
