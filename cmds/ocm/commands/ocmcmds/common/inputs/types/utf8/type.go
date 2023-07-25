// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utf8

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs"
	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
)

const TYPE = "utf8"

func init() {
	inputs.DefaultInputTypeScheme.Register(inputs.NewInputType(TYPE, &Spec{},
		usage, ConfigHandler()))
}

const usage = `
This blob type is used to provide inline text based content (UTF8). The
specification supports the following fields:
- **<code>text</code>** *string*

  The utf8 string content to provide.
` + cpi.ProcessSpecUsage
