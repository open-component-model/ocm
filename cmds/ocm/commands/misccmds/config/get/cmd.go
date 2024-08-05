package get

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/clicfgattr"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/common/options/destoption"
	"ocm.software/ocm/cmds/ocm/commands/misccmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Config
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, output.OutputOptions(outputs), destoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "<options>",
		Short: "Get evaluated config for actual command call",
		Long: `
Evaluate the command line arguments and all explicitly
or implicitly used configuration files and provide
a single configuration object.
`,
	}
}

func (o *Command) AddFlags(set *pflag.FlagSet) {
	o.BaseCommand.AddFlags(set)
}

func (o *Command) Run() error {
	cfg := clicfgattr.Get(o.Context)
	if cfg == nil {
		out.Outf(o.Context, "no configuration found")
		return nil
	}
	opts := output.From(o)

	opts.Output.Add(output.AsManifest(cfg))

	dest := destoption.From(o)
	if dest.Destination != "" {
		file, err := dest.PathFilesystem.OpenFile(dest.Destination, vfs.O_CREATE|vfs.O_TRUNC|vfs.O_WRONLY, 0o600)
		if err != nil {
			return errors.Wrapf(err, "cannot create output file %q", dest.Destination)
		}
		opts.Output.(output.Destination).SetDestination(file)
		defer file.Close()
	}
	err := opts.Output.Out()
	if err == nil && dest.Destination != "" {
		out.Outf(o.Context, "config written to %q\n", dest.Destination)
	}
	return err
}

var outputs = output.NewOutputs(output.DefaultYAMLOutput).AddManifestOutputs()
