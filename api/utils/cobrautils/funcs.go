package cobrautils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"ocm.software/ocm/api/utils/cobrautils/groups"
)

var templatefuncs = map[string]interface{}{
	"useLine":                useLine,
	"commandPath":            commandPath,
	"indent":                 indent,
	"skipCommand":            skipCommand,
	"soleCommand":            soleCommand,
	"title":                  cases.Title(language.English).String,
	"substituteCommandLinks": substituteCommandLinks,
	"flagUsages":             flagUsages,
	"commandList":            commandList,
}

const COMMAND_PATH_SUBSTITUTION = "ocm.software/commandPathSubstitution"

func SetCommandSubstitutionForTree(cmd *cobra.Command, remove int, prepend []string) {
	SetCommandSubstitution(cmd, remove, prepend)
	for _, c := range cmd.Commands() {
		SetCommandSubstitutionForTree(c, remove, prepend)
	}
}

func SetCommandSubstitution(cmd *cobra.Command, remove int, prepend []string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[COMMAND_PATH_SUBSTITUTION] = fmt.Sprintf("%d:%s", remove, strings.Join(prepend, " "))
}

func useLine(c *cobra.Command) string {
	cp := commandPath(c)
	i := strings.Index(c.Use, " ")
	if i > 0 {
		cp += c.Use[i:]
	}
	if !c.DisableFlagsInUseLine && c.HasAvailableFlags() && !strings.Contains(cp, "[flags]") {
		cp += " [flags]"
	}
	return cp
}

func commandPath(c *cobra.Command) string {
	if c.Annotations != nil {
		subst := c.Annotations[COMMAND_PATH_SUBSTITUTION]
		if subst != "" {
			i := strings.Index(subst, ":")
			if i > 0 {
				remove, err := strconv.Atoi(subst[:i])
				if err == nil {
					fields := strings.Split(c.CommandPath(), " ")
					fields = sliceutils.CopyAppend(strings.Split(subst[i+1:], " "), fields[remove:]...)
					return strings.Join(fields, " ")
				}
			}
		}
	}
	return c.CommandPath()
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
