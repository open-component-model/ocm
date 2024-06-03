package logging

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/logging"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	logcfg "github.com/open-component-model/ocm/pkg/contexts/datacontext/config/logging"
	utils2 "github.com/open-component-model/ocm/pkg/listformat"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "logging",
		Short: "Configured logging keys",
		Example: `
    type: ` + logcfg.ConfigType + `
    contextType: ` + datacontext.CONTEXT_TYPE + `
    settings:
      defaultLevel: Info
      rules:
        - ...
`,
		Long: `
Logging can be configured as part of the ocm config file (<CMD>ocm configfile</CMD>)
or by command line options of the <CMD>ocm</CMD> command. Details about
the YAML structure of a logging settings can be found on https://github.com/mandelsoft/logging.

The command line also supports some quick-config options for enabling log levels
for dedicated tags and realms or realm prefixes (logging keys).

` + describe("tags", logging.GetTagDefinitions()) + `

` + describe("realms", logging.GetRealmDefinitions()),
	}
}

func describe(name string, defs logging.Definitions) string {
	if len(defs) == 0 {
		return fmt.Sprintf("There are no defined *%s*.", name)
	}
	return fmt.Sprintf(`The following *%s* are used by the command line tool:
%s
`, name, utils2.FormatMapElements("", defs, func(e []string) string {
		return strings.Join(e, ", ")
	}))
}
