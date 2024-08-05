package upload

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload/put"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/upload/validate"
)

const Name = "upload"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "upload specific operations",
		Long: `
This command group provides all commands used to implement an uploader
described by an uploader descriptor.`,
	}

	cmd.AddCommand(validate.New(p))
	cmd.AddCommand(put.New(p))
	return cmd
}
