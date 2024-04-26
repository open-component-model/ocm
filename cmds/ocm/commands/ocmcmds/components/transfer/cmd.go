package transfer

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/omitaccesstypeoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/skipupdateoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/srcbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/stoponexistingoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/out"
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
$ ocm transfer components -t tgz ghcr.io/mandelsoft/kubelink ctf.tgz
$ ocm transfer components -t tgz --repo OCIRegistry::ghcr.io mandelsoft/kubelink ctf.tgz
`,
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
	err := transfer.TransferVersion(a.printer, a.closure, o.ComponentVersion, a.target, a.handler)
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
		return fmt.Errorf("transfer finished with %d error(s)", a.errors.Len())
	}

	if a.cmd.BOMFile != "" {
		bom := BOM{}
		for _, nv := range generics.KeyList(a.closure) {
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
