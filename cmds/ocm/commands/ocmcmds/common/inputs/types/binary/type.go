package binary

import (
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const (
	TYPE   = "binary"
	TypeV1 = TYPE + runtime.VersionSeparator + "v1"
)

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TypeV1, &Spec{},
		"", ConfigHandler()))
}

const usage = `
This blob type is used to provide base64 encoded binary content. The
specification supports the following fields:
- **<code>data</code>** *[]byte*

  The binary data to provide.
` + cpi.ProcessSpecUsage
