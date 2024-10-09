package git

import (
	"ocm.software/ocm/api/oci/annotations"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "git"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The repository type allows accessing an arbitrary git repository
using the manifest annotation <code>` + annotations.COMPVERS_ANNOTATION + `</code>.
The ref can be used to further specify the branch or tag to checkout, otherwise the remote HEAD is used.

This blob type specification supports the following fields: 
- **<code>repository</code>** *string*

  This REQUIRED property describes the URL of the git repository to access. All git URL formats are supported.

- **<code>ref</code>** *string*

  This OPTIONAL property can be used to specify the remote branch or tag to checkout (commonly called ref). 
  If not set, the default HEAD (remotes/origin/HEAD) of the remote is used.

- **<code>commit</code>** *string*

  This OPTIONAL property can be used to specify the commit hash to checkout.
  If not set, the default HEAD of the ref is used.
`
