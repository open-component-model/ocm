// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "docker"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage))
}

const usage = `
The path must denote an image tag that can be found in the local
docker daemon. The denoted image is packed as OCI artefact set.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the image name to import from the
  local docker daemon.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artefact access. It is prefixed by the component name if
  it does not start with slash "/".`
