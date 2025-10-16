package accessmethod

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/compose"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/get"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/accessmethod/validate"
)

const Name = "accessmethod"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "access method operations",
		Long: `This command group provides all commands used to implement an access method
described by an access method descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(get.New(p))
	cmd.AddCommand(compose.New(p))
	return cmd
}
