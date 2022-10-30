// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cmds

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/cobrautils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/accessmethod"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/info"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/upload"
)

type PluginCommand struct {
	command *cobra.Command
	plugin  ppi.Plugin
}

func NewPluginCommand(p ppi.Plugin) *PluginCommand {
	pcmd := &PluginCommand{
		plugin: p,
	}
	cmd := &cobra.Command{
		Use:                   p.Name() + " <subcommand> <options> <args>",
		Short:                 "OCM plugin " + p.Name(),
		Long:                  "OCM plugin " + p.Name(),
		Version:               p.Version(),
		TraverseChildren:      true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		SilenceErrors:         true,
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cobrautils.TweakCommand(cmd, nil)

	cmd.AddCommand(info.New(p))
	cmd.AddCommand(accessmethod.New(p))
	cmd.AddCommand(upload.New(p))

	cmd.InitDefaultHelpCmd()
	var help *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			help = c
			break
		}
	}
	// help.Use="help <topic>"
	help.DisableFlagsInUseLine = true

	p.Options().AddFlags(cmd.Flags())
	pcmd.command = cmd
	return pcmd
}

type Error struct {
	Error string `json:"error"`
}

func (p *PluginCommand) Execute(args []string) error {
	p.command.SetArgs(args)
	err := p.command.Execute()
	if err != nil {
		result, err2 := json.Marshal(Error{err.Error()})
		if err2 != nil {
			return err2
		}
		p.command.PrintErrln(string(result))
	}
	return err
}
