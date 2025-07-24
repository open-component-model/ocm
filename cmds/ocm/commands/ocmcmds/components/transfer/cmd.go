package transfer

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/compdesc"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/spiff"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/omitaccesstypeoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/skipupdateoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/srcbyvalueoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/stoponexistingoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Components
	Verb  = verbs.Transfer
)

type Command struct {
	utils.BaseCommand

	Refs                []string
	TargetName          string
	BOMFile             string
	DisableBlobHandlers bool
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		versionconstraintsoption.New(),
		repooption.New(),
		formatoption.New(),
		closureoption.New("component reference"),
		lookupoption.New(),
		overwriteoption.New(),
		skipupdateoption.New(),
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
		Use:   "[<options>] {<component-reference>} <target>",
		Args:  cobra.MinimumNArgs(1),
		Short: "transfer component version",
		Long: `
Transfer all component versions specified to the given target repository.
If only a component (instead of a component version) is specified all versions
are transferred.
`,
		Example: `
$ ocm transfer components -t tgz ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0 ./ctf.tgz
$ ocm transfer components --latest -t tgz --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli ./ctf.tgz
$ ocm transfer components --latest --copy-resources --type directory ghcr.io/open-component-model/ocm//ocm.software/ocmcli ./ctf
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.BOMFile, "bom-file", "B", "", "file name to write the component version BOM")
	fs.BoolVarP(&o.DisableBlobHandlers, "disable-uploads", "", false, "disable standard upload handlers for transport")
}

func (o *Command) Complete(args []string) error {
	o.Refs = args[:len(args)-1]
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is required")
	}
	o.TargetName = args[len(args)-1]
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()
	session.Finalize(o.OCMContext())

	if o.DisableBlobHandlers {
		out.Outln(o.Context, "standard blob upload handlers are disabled.")
		o.Context.OCMContext().DisableBlobHandlers()
	}
	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	err = uploaderoption.From(o).Register(o)
	if err != nil {
		return err
	}

	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, formatoption.From(o).ChangedFormat(), o.Context.FileSystem())
	if err != nil {
		return err
	}

	transferopts := &spiff.Options{}
	transferhandler.From(o.ConfigContext(), transferopts)
	transferhandler.ApplyOptions(transferopts, append(options.FindOptions[transferhandler.TransferOption](o),
		spiff.Script(scriptoption.From(o).ScriptData),
		spiff.ScriptFilesystem(o.FileSystem()),
	)...)
	thdlr, err := spiff.New(transferopts)
	if err != nil {
		return err
	}
	hdlr := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, comphdlr.OptionsFor(o))
	err = utils.HandleOutput(&action{
		cmd:     o,
		printer: common.NewPrinter(o.Context.StdOut()),
		target:  target,
		handler: thdlr,
		closure: transfer.TransportClosure{},
		errors:  errors.ErrListf("transfer errors"),
	}, hdlr, utils.StringElemSpecs(o.Refs...)...)
	if err != nil {
		return err
	}
	return session.Close()
}

/////////////////////////////////////////////////////////////////////////////

type action struct {
	cmd     *Command
	printer common.Printer
	target  ocm.Repository
	handler transferhandler.TransferHandler
	closure transfer.TransportClosure
	errors  *errors.ErrorList
}

var _ output.Output = (*action)(nil)

func (a *action) Add(e interface{}) error {
	o, ok := e.(*comphdlr.Object)
	if !ok {
		return fmt.Errorf("object of type %T is not a valid comphdlr.Object", e)
	}
	sub, h, err := a.handler.TransferVersion(o.Repository, nil, compdesc.NewComponentReference("", o.ComponentVersion.GetName(), o.ComponentVersion.GetVersion(), nil), a.target)
	if err != nil {
		return errors.Wrapf(err, "cannot transfer component version %s/%s", o.ComponentVersion.GetName(), o.ComponentVersion.GetVersion())
	}
	if sub == nil {
		return fmt.Errorf("cannot transfer component version %s/%s", o.ComponentVersion.GetName(), o.ComponentVersion.GetVersion())
	}
	err = transfer.TransferVersion(a.printer, a.closure, sub, a.target, h)
	sub.Close()
	a.errors.Add(err)
	if err != nil {
		a.printer.Printf("Error: %s\n", err)
	}
	return nil
}

func (a *action) Close() error {
	return nil
}

func (a *action) Out() error {
	a.printer.Printf("%d versions transferred\n", len(a.closure))
	if a.errors.Result() != nil {
		sum := "Error summary:"
		for _, e := range a.errors.Entries() {
			sum = fmt.Sprintf("%s\n- %s", sum, e)
		}
		return fmt.Errorf("transfer finished with %d error(s)\n%s\n", a.errors.Len(), sum)
	}

	if a.cmd.BOMFile != "" {
		bom := BOM{}
		for _, nv := range maputils.Keys(a.closure, common.CompareNameVersion) {
			bom.List = append(bom.List, BomEntry{
				Component: nv.GetName(),
				Version:   nv.GetVersion(),
			})
		}
		data, err := json.Marshal(&bom)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal BOM")
		}
		err = vfs.WriteFile(a.cmd.FileSystem(), a.cmd.BOMFile, data, 0o640)
		if err != nil {
			return errors.Wrapf(err, "cannot write BOM")
		}
	}
	return nil
}

type BomEntry struct {
	Component string `json:"component"`
	Version   string `json:"version"`
}
type BOM struct {
	List []BomEntry `json:"componentVersions"`
}
