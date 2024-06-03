package docker

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/pkg/contexts/oci/annotations"
)

const TYPE = "docker"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The path must denote an image tag that can be found in the local
docker daemon. The denoted image is packed as OCI artifact set.
The OCI image will contain an informational back link to the component version
using the manifest annotation <code>` + annotations.COMPVERS_ANNOTATION + `</code>.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the image name to import from the
  local docker daemon.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artifact access. It is prefixed by the component name if
  it does not start with slash "/".
`
