package command

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/ppi"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const Name = "command"

func New(p ppi.Plugin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   Name,
		Short: "CLI command extensions",
		Long: `This command group provides all CLI command extensions
described by an access method descriptor (<CMD>` + p.Name() + ` descriptor</CMD>.`,
	}

	octx := ocm.DefaultContext()
	ctx := octx.BindTo(context.Background())
	for _, n := range p.Commands() {
		c := n.Command()

		nested := c.PreRunE
		c.PreRunE = func(cmd *cobra.Command, args []string) error {
			c.SetContext(ctx)
			err := ConfigureFromStdIn(octx)
			if err != nil {
				return err
			}
			if nested != nil {
				return nested(cmd, args)
			}
			return nil
		}
		cmd.AddCommand(n.Command())
	}
	return cmd
}

func ConfigureFromStdIn(ctx ocm.Context) error {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(raw)) == "" {
		return nil
	}
	_, err = ctx.ConfigContext().ApplyData(raw, runtime.DefaultYAMLEncoding, " cli config")
	// Ugly, enforce configuration nupdate
	ctx.GetResolver()
	return err
}
