package input

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/compose"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/get"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/input/validate"
)

const Name = "input"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "input operations for component version composition",
		Long: `This command group provides all commands used to implement an input type
described by an input type descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(get.New(p))
	cmd.AddCommand(compose.New(p))
	return cmd
}
