// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
)

const TYPE = "helm"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage, ConfigHandler()))
}

const usage = `
The path must denote an helm chart archive or directory
relative to the resources file.
The denoted chart is packed as an OCI artefact set.
Additional provider info is taken from a file with the same name
and the suffix <code>.prov</code>.

If the chart should just be stored as archive, please use the 
type <code>file</code> or <code>dir</code>.

This blob type specification supports the following fields: 
- **<code>path</code>** *string*

  This REQUIRED property describes the file path to the helm chart relative to the
  resource file location.

- **<code>version</code>** *string*

  This OPTIONAL property can be set to configure an explicit version hint.
  If not specified the versio from the chart will be used.
  Basically, it is a good practice to use the component version for local resources
  This can be achieved by using templating for this attribute in the resource file.
`
