package attributes

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext"
)

func New(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "attributes",
		Short: "configuration attributes used to control the behaviour",
		Long: `
The OCM library supports a set of attributes, which can be used to influence
the behaviour of various functions. The CLI also supports setting of those
attributes using the config file (see <CMD>ocm configfile</CMD>) or by
command line options of the main command (see <CMD>ocm</CMD>).

The following options are available in the currently used version of the
OCM library:
` + Attributes(),
	}
}

func Attributes() string {
	s := ""
	sep := ""
	for _, a := range datacontext.DefaultAttributeScheme.KnownTypeNames() {
		t, err := datacontext.DefaultAttributeScheme.GetType(a)
		if err != nil {
			continue
		}

		desc := t.Description()
		if !strings.Contains(desc, "not via command line") {
			for strings.HasPrefix(desc, "\n") {
				desc = desc[1:]
			}
			for strings.HasSuffix(desc, "\n") {
				desc = desc[:len(desc)-1]
			}
			lines := strings.Split(desc, "\n")
			title := lines[0]
			desc = "  " + strings.Join(lines[1:], "\n  ")
			short := ""
			for k, v := range datacontext.DefaultAttributeScheme.Shortcuts() {
				if v == a {
					short = short + ",<code>" + k + "</code>"
				}
			}
			if len(short) > 0 {
				short = " [" + short[1:] + "]"
			}
			s = fmt.Sprintf("%s%s- <code>%s</code>%s: %s\n\n%s", s, sep, a, short, title, desc)
			sep = "\n\n"
		}
	}
	return s
}
