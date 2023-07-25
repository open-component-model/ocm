// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package dockermulti

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "dockermulti"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
This input type describes the composition of a multi-platform OCI image.
The various variants are taken from the local docker daemon. They should be 
built with the buildx command for cross platform docker builds.
The denoted images, as well as the wrapping image index is packed as OCI artifact set.

This blob type specification supports the following fields:
- **<code>variants</code>** *[]string*

  This REQUIRED property describes a set of  image names to import from the
  local docker daemon used to compose a resulting image index.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artifact access. It is prefixed by the component name if
  it does not start with slash "/".
`
