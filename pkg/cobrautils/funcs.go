package cobrautils

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
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
	"commandList":            commandList,
}

func flagUsages(fs *pflag.FlagSet) string {
	return groups.FlagUsagesWrapped(fs, 0)
}

func commandList(cmds []*cobra.Command) string {
	if len(cmds) == 0 {
		return ""
	}
	var list []string
	for _, c := range cmds {
		if c.IsAvailableCommand() {
			list = append(list, c.Name())
		}
	}
	if len(list) == 0 {
		return ""
	}
	sort.Strings(list)
	return "[" + strings.Join(list, ", ") + "]"
}

func substituteCommandLinks(desc string) string {
	_, desc = SubstituteCommandLinks(desc, func(pname string) string {
		return "\u00ab" + pname + "\u00bb"
	})
	return desc
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
