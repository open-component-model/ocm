package binary

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const TYPE = "binary"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
}

const usage = `
This blob type is used to provide base64 encoded binary content. The
specification supports the following fields:
- **<code>data</code>** *[]byte*

  The binary data to provide.
` + cpi.ProcessSpecUsage
