package command

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi/cmds/common"
)

const (
	Name         = "command"
	OptCliConfig = common.OptCliConfig
)

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "CLI command extensions",
		Long: `This command group provides all CLI command extensions
described by an access method descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
		TraverseChildren: true,
	}
	var cliconfig string
	cmd.Flags().StringVarP(&cliconfig, OptCliConfig, "", "", "path to cli configuration file")

	for _, n := range p.Commands() {
		c := n.Command()
		c.TraverseChildren = true

		nested := c.PreRunE
		c.PreRunE = func(cmd *cobra.Command, args []string) error {
			ctx, err := ConfigureFromFile(context.Background(), cliconfig)
			if err != nil {
				return err
			}
			c.SetContext(ctx)
			if nested != nil {
				return nested(cmd, args)
			}
			return nil
		}
		cmd.AddCommand(n.Command())
	}
	return cmd
}

func ConfigureFromFile(ctx context.Context, path string) (context.Context, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ctx, err
	}

	if handler != nil {
		return handler.HandleConfig(ctx, data)
	}
	return ctx, nil
}
