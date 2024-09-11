package describe

import (
	"os"

	"github.com/spf13/cobra"

	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/ocm/plugin/common"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/misc"
)

const NAME = "describe"

func New(p ppi.Plugin) *cobra.Command {
	return &cobra.Command{
		Use:   NAME,
		Short: "describe plugin",
		Long:  "Display a detailed description of the capabilities of this OCM plugin.",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			d := p.Descriptor()
			common.DescribePluginDescriptor(action.DefaultRegistry(), &d, misc.NewPrinter(os.Stdout))
			return nil
		},
	}
}
