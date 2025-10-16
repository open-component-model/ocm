package transfer

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/spiff"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/omitaccesstypeoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/skipupdateoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/srcbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/stoponexistingoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.CommonTransportArchive
	Verb  = verbs.Transfer
)

type Command struct {
	utils.BaseCommand

	SourceName string
	TargetName string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		closureoption.New("component reference"),
		skipupdateoption.New(),
		overwriteoption.New(),
		lookupoption.New(),
		formatoption.New(),
		rscbyvalueoption.New(),
		srcbyvalueoption.New(),
		omitaccesstypeoption.New(),
		stoponexistingoption.New(),
		uploaderoption.New(ctx.OCMContext()),
		scriptoption.New(),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <ctf> <target>",
		Args:  cobra.MinimumNArgs(2),
		Short: "transfer transport archive",
		Long: `
Transfer content of a Common Transport Archive to the given target repository.
`,
		Example: `
$ ocm transfer ctf ctf.tgz ghcr.io/mandelsoft/components
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Complete(args []string) error {
	o.SourceName = args[0]
	o.TargetName = args[1]
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	err = uploaderoption.From(o).Register(o)
	if err != nil {
		return err
	}

	src, err := ctf.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.SourceName, 0, o.FileSystem())
	if err != nil {
		return errors.Wrapf(err, "cannot open source")
	}
	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, formatoption.From(o).ChangedFormat(), o.Context.FileSystem())
	if err != nil {
		return err
	}

	thdlr, err := spiff.New(
		append(options.FindOptions[transferhandler.TransferOption](o),
			spiff.Script(scriptoption.From(o).ScriptData),
			spiff.ScriptFilesystem(o.FileSystem()),
		)...)
	if err != nil {
		return err
	}
	a := &action{
		printer: common.NewPrinter(o.Context.StdOut()),
		target:  target,
		handler: thdlr,
		closure: transfer.TransportClosure{},
		errors:  errors.ErrListf("transfer errors"),
	}
	return a.Execute(src)
}

/////////////////////////////////////////////////////////////////////////////

type action struct {
	printer common.Printer
	target  ocm.Repository
	handler transferhandler.TransferHandler
	closure transfer.TransportClosure
	errors  *errors.ErrorList
}

func (a *action) Execute(src ocm.Repository) error {
	return transfer.TransferComponents(a.printer, a.closure, src, "", true, a.target, a.handler)
}
