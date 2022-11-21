// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"github.com/spf13/cobra"

	handler "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	common2 "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/cobrautils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
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
	common2.DescribePlugin(p, out)
	if a.Count > 1 {
		a.Printer.Printf("----------------------\n")
	}
	a.Printer.Printf("%s\n", cobrautils.CleanMarkdown(buf.String()))
	return nil
}

func (a *action) Close() error {
	return nil
}

func (a *action) Out() error {
	a.Printer.Printf("*** found %d plugins\n", a.Count)
	return nil
}
