package utf8

import (
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const (
	TYPE          = "utf8"
	TypeV1        = TYPE + runtime.VersionSeparator + "v1"
	UPPER_TYPE    = "UTF8"
	UPPER_TYPE_V1 = UPPER_TYPE + runtime.VersionSeparator + "v1"
)

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TypeV1, &Spec{},
		"", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE, &Spec{},
		"", ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(UPPER_TYPE_V1, &Spec{},
		"", ConfigHandler()))
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
