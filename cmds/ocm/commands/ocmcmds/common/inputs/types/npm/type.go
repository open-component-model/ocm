package npm

import (
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "npm"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The <code>registry</code> is the url pointing to the npm registry from which a resource is 
downloaded. 

This blob type specification supports the following fields:
- **<code>registry</code>** *string*

  This REQUIRED property describes the url from which the resource is to be
  downloaded.

- **<code>package</code>** *string*
	
  This REQUIRED property describes the name of the package to download.

- **<code>version</code>** *string*

  This is an OPTIONAL property describing the version of the package to download. If
  not defined, latest will be used automatically.
`
