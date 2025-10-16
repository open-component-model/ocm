package command

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/cobrautils"
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

	found := false
	for _, n := range p.Commands() {
		found = true
		c := n.Command()
		c.TraverseChildren = true

		nested := c.PreRunE
		c.PreRunE = func(cmd *cobra.Command, args []string) error {
			var err error

			ctx := context.Background()
			if cliconfig != "" {
				ctx, err = ConfigureFromFile(ctx, cliconfig)
				if err != nil {
					return err
				}
			}
			c.SetContext(ctx)
			if nested != nil {
				return nested(cmd, args)
			}
			return nil
		}
		cmd.AddCommand(n.Command())
	}
	if found {
		cobrautils.TweakHelpCommandFor(cmd)
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
