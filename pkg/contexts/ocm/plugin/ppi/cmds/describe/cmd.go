package describe

import (
	"os"

	"github.com/spf13/cobra"

	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/common"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
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
			common.DescribePluginDescriptor(action.DefaultRegistry(), &d, common2.NewPrinter(os.Stdout))
			return nil
		},
	}
}
