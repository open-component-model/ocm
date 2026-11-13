package groups

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"ocm.software/ocm/api/utils"
)

const FlagGroupAnnotation = "flag-group-annotation"

// boolFlag mirrors pflag's unexported interface for flags that need no value.
type boolFlag interface {
	pflag.Value
	IsBoolFlag() bool
}

// plainValue hides IsBoolFlag, so pflag (>= v1.1.0) does not render the
// "[=true|false]" suffix for bool flags and help output keeps the
// established "--name   usage" format.
type plainValue struct{ pflag.Value }

// usageFlagSet returns a copy of f, used for rendering only, whose bool
// flags are wrapped into plainValue. f itself is not modified.
func usageFlagSet(f *pflag.FlagSet) *pflag.FlagSet {
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.SortFlags = f.SortFlags
	f.VisitAll(func(flag *pflag.Flag) {
		c := *flag
		if b, ok := flag.Value.(boolFlag); ok && b.IsBoolFlag() && b.Type() == "bool" {
			c.Value = plainValue{flag.Value}
		}
		fs.AddFlag(&c)
	})
	return fs
}

// FlagUsagesWrapped returns a string containing the usage information
// for all flags in the FlagSet. Wrapped to `cols` columns (0 for no
// wrapping)
// It groups flags according to group annotation.
func FlagUsagesWrapped(f *pflag.FlagSet, cols int) string {
	lines := DetermineGroups(f, cols)

	sep := ""
	buf := new(bytes.Buffer)
	for _, g := range utils.StringMapKeys(lines) {
		if g != "" {
			fmt.Fprintln(buf, sep+"  "+g+":")
		}
		sep = "\n"
		for _, line := range lines[g] {
			fmt.Fprintln(buf, line)
		}
	}

	return buf.String()
}

type UsageGroup struct {
	Title  string
	Usages string
}

func GroupedFlagUsagesWrapped(f *pflag.FlagSet, cols int) []UsageGroup {
	lines := DetermineGroups(f, cols)

	var groups []UsageGroup
	for _, g := range utils.StringMapKeys(lines) {
		buf := new(bytes.Buffer)
		for _, line := range lines[g] {
			fmt.Fprintln(buf, line)
		}
		groups = append(groups, UsageGroup{
			Title:  g,
			Usages: buf.String(),
		})
	}

	return groups
}

func DetermineGroups(f *pflag.FlagSet, cols int) map[string][]string {
	lines := map[string][]string{}
	for _, line := range strings.Split(usageFlagSet(f).FlagUsagesWrapped(cols), "\n") {
		i := strings.Index(line, "--")
		if i < 0 {
			continue
		}

		name := line[i+2:]
		if end := strings.IndexAny(name, " ["); end >= 0 {
			name = name[:end]
		}
		groups := []string{""}
		if flag := f.Lookup(name); flag != nil && flag.Annotations != nil {
			g := flag.Annotations[FlagGroupAnnotation]
			if len(g) > 0 {
				groups = g
			}
		}
		for _, g := range groups {
			lines[g] = append(lines[g], line)
		}
	}
	return lines
}
