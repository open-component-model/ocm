// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobrautils

import (
	"strings"

	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/open-component-model/ocm/pkg/cobrautils/groups"
)

var templatefuncs = map[string]interface{}{
	"indent":                 indent,
	"skipCommand":            skipCommand,
	"soleCommand":            soleCommand,
	"title":                  cases.Title(language.English).String,
	"substituteCommandLinks": substituteCommandLinks,
	"flagUsages":             flagUsages,
}

func flagUsages(fs *pflag.FlagSet) string {
	return groups.FlagUsagesWrapped(fs, 0)
}

func substituteCommandLinks(desc string) string {
	return SubstituteCommandLinks(desc, func(pname string) string {
		return "\u00ab" + pname + "\u00bb"
	})
}

func soleCommand(s string) bool {
	return !strings.Contains(s, " ")
}

func skipCommand(s string) string {
	i := strings.Index(s, " ")
	if i < 0 {
		return ""
	}
	for ; i < len(s); i++ {
		if s[i] != ' ' {
			return s[i:]
		}
	}
	return ""
}

func indent(n int, s string) string {
	gap := ""
	for ; n > 0; n-- {
		gap += " "
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if len(l) > 0 {
			lines[i] = gap + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
