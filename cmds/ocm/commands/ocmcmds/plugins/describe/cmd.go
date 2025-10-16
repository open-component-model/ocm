package describe

import (
	"strings"

	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/cobrautils"
	common "ocm.software/ocm/api/utils/misc"
	handler "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Plugins
	Verb  = verbs.Describe
)

type Command struct {
	utils.BaseCommand

	Names []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<plugin name>}",
		Short: "get plugins",
		Long: `
Describes provides comprehensive information about the capabilities of
a plugin.
`,
		Example: `
$ ocm describe plugins
$ ocm describe plugins demo
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Complete(args []string) error {
	o.Names = args
	return nil
}

func (o *Command) Run() error {
	hdlr := handler.NewTypeHandler(o.Context.OCM())
	return utils.HandleOutput(NewAction(o), hdlr, utils.StringElemSpecs(o.Names...)...)
}

/////////////////////////////////////////////////////////////////////////////

type action struct {
	Printer common.Printer
	Count   int
}

func NewAction(o *Command) *action {
	return &action{
		Printer: common.NewPrinter(o.StdOut()),
	}
}

func (a *action) Add(e interface{}) error {
	a.Count++
	p := handler.Elem(e)

	out, buf := common.NewBufferedPrinter()
	DescribePlugin(p, out)
	if a.Count > 1 {
		a.Printer.Printf("----------------------\n")
	}
	desc := cobrautils.CleanMarkdown(buf.String())
	if !strings.HasSuffix(desc, "\n") {
		desc += "\n"
	}
	a.Printer.Printf("%s", desc)
	return nil
}

func (a *action) Close() error {
	return nil
}

func (a *action) Out() error {
	a.Printer.Printf("*** found %d plugins\n", a.Count)
	return nil
}
