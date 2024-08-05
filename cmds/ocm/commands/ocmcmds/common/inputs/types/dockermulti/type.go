package dockermulti

import (
	"ocm.software/ocm/api/oci/annotations"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "dockermulti"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
This input type describes the composition of a multi-platform OCI image.
The various variants are taken from the local docker daemon. They should be 
built with the "buildx" command for cross platform docker builds (see https://ocm.software/docs/tutorials/best-practices/#building-multi-architecture-images).
The denoted images, as well as the wrapping image index, are packed as OCI
artifact set.
They will contain an informational back link to the component version
using the manifest annotation <code>` + annotations.COMPVERS_ANNOTATION + `</code>.

This blob type specification supports the following fields:
- **<code>variants</code>** *[]string*

  This REQUIRED property describes a set of  image names to import from the
  local docker daemon used to compose a resulting image index.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artifact access. It is prefixed by the component name if
  it does not start with slash "/".
`
