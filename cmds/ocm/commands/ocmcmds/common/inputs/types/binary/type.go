// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package binary

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const TYPE = "binary"

func init() {
	inputs.DefaultInputTypeScheme.Register(TYPE, inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
}

const usage = `
The content is compressed if the <code>compress</code> field
is set to <code>true</code>.

This blob type specification supports the following fields:
- **<code>data</code>** *[]byte*

  The binary data to provide.
` + cpi.ProcessSpecUsage
