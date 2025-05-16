package routingslip

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/stringutils"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
)

func EntryUsage(scheme EntryTypeScheme, cli bool) string {
	s := `
The following list describes the well-known entry types explicitly supported
by this version of the CLI, their versions and specification formats. Other
kinds of entries can be configured using the <code>--entry</code> option.
`
	type method struct {
		desc     string
		versions map[string]string
		options  flagsets.ConfigOptionTypeSetHandler
	}

	descs := map[string]*method{}

	// gather info for kinds and versions
	for _, n := range scheme.KnownTypeNames() {
		kind, vers := runtime.KindVersion(n)

		info := descs[kind]
		if info == nil {
			info = &method{versions: map[string]string{}}
			descs[kind] = info
		}

		if vers == "" {
			vers = "v1"
		}
		if _, ok := info.versions[vers]; !ok {
			info.versions[vers] = ""
		}

		t := scheme.GetType(n)

		if t.ConfigOptionTypeSetHandler() != nil {
			info.options = t.ConfigOptionTypeSetHandler()
		}
		desc := t.Description()
		if desc != "" {
			info.desc = desc
		}

		desc = t.Format()
		if desc != "" {
			info.versions[vers] = desc
		}
	}

	for _, t := range maputils.OrderedKeys(descs) {
		info := descs[t]
		desc := strings.Trim(info.desc, "\n")
		if desc != "" {
			s = fmt.Sprintf("%s\n- Entry type <code>%s</code>\n\n%s\n\n", s, t, stringutils.IndentLines(desc, "  "))

			format := ""
			for _, f := range maputils.OrderedKeys(info.versions) {
				desc = strings.Trim(info.versions[f], "\n")
				if desc != "" {
					format = fmt.Sprintf("%s\n- Version <code>%s</code>\n\n%s\n", format, f, stringutils.IndentLines(desc, "  "))
				}
			}
			if format != "" {
				s += fmt.Sprintf("  The following versions are supported:\n%s\n", strings.Trim(stringutils.IndentLines(format, "  "), "\n"))
			}
		}
		s += stringutils.IndentLines(flagsets.FormatConfigOptions(info.options), "  ")
	}
	return s
}
