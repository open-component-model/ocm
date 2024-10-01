package signing

import (
	"github.com/spf13/cobra"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/signing/consumer"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/signing/sign"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/signing/verify"
)

const Name = "signing"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "signing related operations",
		Long: `This command group provides all commands used for the signing extension
described by signing descriptor (<CMD>` + p.Name() + ` descriptor</CMD>).`,
	}

	cmd.AddCommand(consumer.New(p))
	cmd.AddCommand(sign.New(p))
	cmd.AddCommand(verify.New(p))
	return cmd
}
