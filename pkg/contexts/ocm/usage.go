//  SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
//  SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"fmt"
	"strings"

	"github.com/open-component-model/ocm/pkg/utils"
)

func AccessUsage(scheme AccessTypeScheme, cli bool) string {
	s := `
The following access methods are known by the system.
Typically there is special support for the CLI artifact add commands.
The following types (with the field <code>type</code> in the <code>access</code> field
are handled:

`

	names := map[string][]string{}
	descs := map[string]string{}

	for _, t := range scheme.KnownTypeNames() {
		base := t
		i := strings.Index(t, "/")
		if i > 0 {
			base = t[:i]
			vers := t[i+1:]
			names[base] = append(names[base], vers)
		}
		desc := scheme.GetAccessType(t).Description(cli)
		if desc != "" {
			descs[base] = desc
		}
	}
	for _, t := range scheme.KnownTypeNames() {
		desc := descs[t]
		if desc != "" {
			s = fmt.Sprintf("%s\n\n- Access type <code>%s</code>\n\n%s", s, t, utils.IndentLines(desc, "  "))
		}
	}
	return s + "\n"
}
