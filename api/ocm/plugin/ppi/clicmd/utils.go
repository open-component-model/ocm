package clicmd

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/spf13/cobra"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	_ "ocm.software/ocm/cmds/ocm/clippi/config"
)

////////////////////////////////////////////////////////////////////////////////

type CobraCommand struct {
	cmd               *cobra.Command
	verb              string
	realm             string
	objname           string
	cliConfigRequired bool
}

var _ ppi.Command = (*CobraCommand)(nil)

// NewCLICommand created a CLI command based on a preconfigured cobra.Command.
// Optionally, a verb can be specified. If given additionally a realm
// can be given.
// verb and realm are used to add the command at the appropriate places in
// the command hierarchy of the ocm CLI.
// If nothing is specified, the command will be a new top-level command.
// To access the configured ocm context use the Context attribute
// of the cobra command. The ocm context is bound to it.
//
//	ocm.FromContext(cmd.Context())
func NewCLICommand(cmd *cobra.Command, opts ...Option) (ppi.Command, error) {
	eff := optionutils.EvalOptions(opts...)
	if eff.Verb == "" && eff.Realm != "" {
		return nil, errors.New("realm without verb not allowed")
	}
	cmd.DisableFlagsInUseLine = true
	return &CobraCommand{cmd, eff.Verb, eff.Realm, eff.ObjectType, optionutils.AsBool(eff.RequireCLIConfig, false)}, nil
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

func (c *CobraCommand) ObjectType() string {
	if c.objname == "" {
		return c.Name()
	}
	return c.objname
}

func (c *CobraCommand) Verb() string {
	return c.verb
}

func (c *CobraCommand) Realm() string {
	return c.realm
}

func (c *CobraCommand) CLIConfigRequired() bool {
	return c.cliConfigRequired
}

func (c *CobraCommand) Command() *cobra.Command {
	return c.cmd
}
