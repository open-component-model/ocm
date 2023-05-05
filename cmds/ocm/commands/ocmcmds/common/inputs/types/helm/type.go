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
relative to the resources file or a chart name in a helm chart repository.
The denoted chart is packed as an OCI artifact set.
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
  If not specified the version from the chart will be used.
  Basically, it is a good practice to use the component version for local resources
  This can be achieved by using templating for this attribute in the resource file.

- **<code>repository</code>** *string*

  This OPTIONAL property can be used to specify the repository hint for the
  generated local artifact access. It is prefixed by the component name if
  it does not start with slash "/".

- **<code>caCertFile</code>** *string*

  This OPTIONAL property can be used to specify a relative filename for
  the TLS root certificate used to access a helm repository.

- **<code>caCert</code>** *string*

  This OPTIONAL property can be used to specify a TLS root certificate used to
  access a helm repository.
`
