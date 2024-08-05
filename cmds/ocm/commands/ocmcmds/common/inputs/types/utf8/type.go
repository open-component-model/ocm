package utf8

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const TYPE = "utf8"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
}

const usage = `
This blob type is used to provide inline text based content (UTF8). The
specification supports the following fields:
- **<code>text</code>** *string*

  The utf8 string content to provide.

- **<code>json</code>** *JSON or JSON string interpreted as JSON*

  The content emitted as JSON.

- **<code>formattedJson</code>** *YAML/JSON or JSON/YAML string interpreted as JSON*

  The content emitted as formatted JSON.

- **<code>yaml</code>** *AML/JSON or JSON/YAML string interpreted as YAML*

  The content emitted as YAML.
` + cpi.ProcessSpecUsage
