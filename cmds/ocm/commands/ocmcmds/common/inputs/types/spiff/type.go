// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spiff

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/file"
)

const TYPE = "spiff"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{}, usage(), ConfigHandler()))
}

func usage() string {
	return file.Usage("The path must denote a [spiff](https://github.com/mandelsoft/spiff) template relative the the resources file.") + `
- **<code>values</code>** *map[string]any*

  This OPTIONAL property describes an additioanl value binding for the template processing. It will be available
  under the node <code>values</code>.

- **<code>libraries</code>** *[]string*

  This OPTIONAL property describes a list of spiff libraries to include in template
  processing.
`
}
