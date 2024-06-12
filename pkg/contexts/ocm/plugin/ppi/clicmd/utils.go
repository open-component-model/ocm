package clicmd

import (
	_ "github.com/open-component-model/ocm/cmds/ocm/clippi/config"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
)

////////////////////////////////////////////////////////////////////////////////

type CobraCommand struct {
	cmd   *cobra.Command
	verb  string
	realm string
}

var _ ppi.Command = (*CobraCommand)(nil)

// NewCLICommand created a CLI command based on a preconfigured cobra.Command.
// Optionally, a verb can be specified. If given additionally a realm
// can be given.
// verb and realm are used to add the command at the appropriate places in
// the command hierarchy of the ocm CLI.
// If nothing is specified, the command will be a new top-level command.
func NewCLICommand(cmd *cobra.Command, args ...string) ppi.Command {
	verb := ""
	realm := ""

	if len(args) > 0 {
		verb = args[0]
	}
	if len(args) > 0 {
		realm = args[1]
	}
	return &CobraCommand{cmd, verb, realm}
}

func (c *CobraCommand) Name() string {
	return c.cmd.Name()
}

func (c *CobraCommand) Description() string {
	return c.cmd.Long
}

func (c *CobraCommand) Usage() string {
	return c.cmd.Use
}

func (c *CobraCommand) Short() string {
	return c.cmd.Short
}

func (c *CobraCommand) Example() string {
	return c.cmd.Example
}

func (c *CobraCommand) Verb() string {
	return c.verb
}

func (c *CobraCommand) Realm() string {
	return c.realm
}

func (c *CobraCommand) Command() *cobra.Command {
	return c.cmd
}
