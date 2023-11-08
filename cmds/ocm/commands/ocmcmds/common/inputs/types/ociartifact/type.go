// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ociartifact

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "ociArtifact"
const LEGACY_TYPE = "ociImage"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(LEGACY_TYPE, &Spec{}, legacy_usage, ConfigHandler()))
}

const legacy_usage = `
DEPRECATED: This type is deprecated, please use ` + TYPE + ` instead.
`

const usage = `
The path must denote an OCI image reference.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the OVI image reference of the image to
  import.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artifact access. It is prefixed by the component name if
  it does not start with slash "/".
`
