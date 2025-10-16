package action

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/action/execute"
)

const Name = "action"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "action operations",
		Long:  `This command group provides all commands used to implement an action.`,
	}

	cmd.AddCommand(execute.New(p))
	return cmd
}
