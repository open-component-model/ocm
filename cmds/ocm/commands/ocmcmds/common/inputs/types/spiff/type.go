package spiff

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
)

const TYPE = "spiff"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage(), ConfigHandler()))
}

func usage() string {
	return file.Usage("The path must denote a [spiff](https://github.com/mandelsoft/spiff) template relative the resources file.") + `
- **<code>values</code>** *map[string]any*

  This OPTIONAL property describes an additional value binding for the template processing. It will be available
  under the node <code>inputvalues</code>.

- **<code>libraries</code>** *[]string*

  This OPTIONAL property describes a list of spiff libraries to include in template
  processing.

The variable settings from the command line are available as binding, also. They are provided under the node
<code>values</code>.
`
}
