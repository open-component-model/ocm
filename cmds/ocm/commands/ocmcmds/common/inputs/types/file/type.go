package file

import (
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const (
	TYPE          = "file"
	TypeV1        = TYPE + runtime.VersionSeparator + "v1"
	UPPER_TYPE    = "File"
	UPPER_TYPE_V1 = UPPER_TYPE + runtime.VersionSeparator + "v1"
)

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		Usage("The path must denote a file relative the resources file. "), ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TypeV1, &Spec{},
		"", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE, &Spec{},
		"", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE_V1, &Spec{},
		"", ConfigHandler()))
}
