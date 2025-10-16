package mergehandler

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/mergehandler/execute"
)

const Name = "valuemergehandler"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "value merge handler operations",
		Long:  `This command group provides all commands used to implement an value merge handlers.`,
	}

	cmd.AddCommand(execute.New(p))
	return cmd
}
