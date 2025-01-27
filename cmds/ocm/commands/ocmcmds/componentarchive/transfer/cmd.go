package transfer

import (
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/skipupdateoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/srcbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
var (
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	Names = names.ComponentArchive
	Verb  = verbs.Transfer
)

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
type Command struct {
	utils.BaseCommand
	Path       string
	TargetName string
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
// NewCommand creates a new transfer command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{BaseCommand: utils.NewBaseCommand(ctx,
			formatoption.New(),
			lookupoption.New(),
			skipupdateoption.New(),
			overwriteoption.New(),
			rscbyvalueoption.New(),
			srcbyvalueoption.New(),
		)}, utils.Names(Names, names...)...)
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <source> <target>",
		Args:  cobra.MinimumNArgs(2),
		Short: "(DEPRECATED) - Please use " + names.CommonTransportArchive[0] + " instead",
		// this removes the command from the help output - https://github.com/open-component-model/ocm/issues/1242#issuecomment-2609312927
		// Deprecated: "Deprecated - use " + ocm.CommonTransportFormat + " instead",
		Long: `
Transfer a component archive to some component repository. This might
be a CTF Archive or a regular repository.
If the type CTF is specified the target must already exist, if CTF flavor
is specified it will be created if it does not exist.

Besides those explicitly known types a complete repository spec might be configured,
either via inline argument or command configuration file and name.
`,
	}
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) Complete(args []string) error {
	o.Path = args[0]
	o.TargetName = args[1]

	return nil
}

// Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()
	session.Finalize(o.OCMContext())

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	source, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.Path, 0, o.Context)
	if err != nil {
		return err
	}
	session.Closer(source)

	format := formatoption.From(o).ChangedFormat()
	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, format, o.Context.FileSystem())
	if err != nil {
		return err
	}

	transferopts := &standard.Options{}
	transferhandler.From(o.ConfigContext(), transferopts)
	transferhandler.ApplyOptions(transferopts, options.FindOptions[transferhandler.TransferOption](o)...)
	thdlr, err := standard.New(transferopts)
	if err != nil {
		return err
	}
	return transfer.TransferVersion(common.NewPrinter(o.Context.StdOut()), nil, source, target, thdlr)
}
