package references

import (
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/references/add"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/references/get"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var Names = names.References

// NewCommand creates a new command.
func NewCommand(ctx clictx.Context) *cobra.Command {
	cmd := utils.MassageCommand(&cobra.Command{
		Short: "Commands related to component references in component versions",
	}, Names...)
	cmd.AddCommand(get.NewCommand(ctx, get.Verb))
	cmd.AddCommand(add.NewCommand(ctx, add.Verb))
	return cmd
}
