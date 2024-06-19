package group

import (
	"github.com/open-component-model/ocm/cmds/subcmdplugin/cmds/demo"
	"github.com/spf13/cobra"
)

const Name = "group"

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name + " <options>",
		Short: "a provided command group",
		Long:  "A provided command group with a demo command",
	}

	cmd.AddCommand(demo.New())
	return cmd
}
