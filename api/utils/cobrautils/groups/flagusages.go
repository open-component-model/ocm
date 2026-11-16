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
	// currentGroups holds the group membership of the most recently seen flag
	// header. Continuation lines (wrapped or explicitly multi-line usages) have
	// no flag header and are attributed to the same groups as that header.
	// It starts nil so an empty FlagSet yields an empty map rather than a
	// phantom default group.
	var currentGroups []string
	rendered := strings.TrimSuffix(usageFlagSet(f).FlagUsagesWrapped(cols), "\n")
	if rendered == "" {
		return lines
	}
	for _, line := range strings.Split(rendered, "\n") {
		// pflag renders a flag header as either "      --name" or
		// "  -x, --name"; its continuation lines are indented farther right and
		// never carry "--" at these fixed offsets. Only treat such headers as a
		// new flag; everything else continues the previous flag's groups.
		if isFlagHeader(line) {
			name := line[8:]
			if end := strings.IndexAny(name, " ["); end >= 0 {
				name = name[:end]
			}
			currentGroups = []string{""}
			if flag := f.Lookup(name); flag != nil && flag.Annotations != nil {
				g := flag.Annotations[FlagGroupAnnotation]
				if len(g) > 0 {
					currentGroups = g
				}
			}
		}
		for _, g := range currentGroups {
			lines[g] = append(lines[g], line)
		}
	}
	return lines
}

// isFlagHeader reports whether line is a pflag flag header line, i.e. it starts
// a new flag rather than continuing the previous flag's usage text. pflag emits
// "      --name ..." (no shorthand) or "  -x, --name ..." (with shorthand).
func isFlagHeader(line string) bool {
	if len(line) < 8 || line[6:8] != "--" {
		return false
	}
	return line[:6] == "      " || (line[:3] == "  -" && line[4:6] == ", ")
}
