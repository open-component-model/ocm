package download

import (
	"path"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/cmds/ocm/commands/common/options/destoption"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Components
	Verb  = verbs.Download
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new download command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		repooption.New(),
		destoption.New(),
		formatoption.New(),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<components>} ",
		Args:  cobra.MinimumNArgs(1),
		Short: "download ocm component versions",
		Long: `
Download component versions from an OCM repository. The result is stored in
component archives.

The files are named according to the component version name.
`,
	}
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Refs = args
	return err
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o.Context, session))
	if err != nil {
		return err
	}

	hdlr := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository)
	return utils.HandleOutput(&action{cmd: o}, hdlr, utils.StringElemSpecs(o.Refs...)...)
}

////////////////////////////////////////////////////////////////////////////////

type action struct {
	data comphdlr.Objects
	cmd  *Command
}

var _ output.Output = (*action)(nil)

func (d *action) Add(e interface{}) error {
	d.data = append(d.data, e.(*comphdlr.Object))
	return nil
}

func (d *action) Close() error {
	return nil
}

func (d *action) Out() error {
	list := errors.ErrListf("downloading component versions")
	dest := destoption.From(d.cmd)
	format := formatoption.From(d.cmd)
	switch len(d.data) {
	case 0:
		out.Outf(d.cmd.Context, "no component versions found\n")
	case 1:
		f := dest.Destination
		if f == "" {
			f = DefaultFileName(d.data[0]) + format.Format.Suffix()
		}
		return d.Save(d.data[0], f)
	default:
		for _, e := range d.data {
			f := DefaultFileName(e) + format.Format.Suffix()
			err := d.Save(e, f)
			if err != nil {
				list.Add(err)
				out.Outf(d.cmd.Context, "%s failed: %s\n", f, err)
			}
		}
	}
	return list.Result()
}

func (d *action) Save(o *comphdlr.Object, f string) (err error) {
	dest := destoption.From(d.cmd)
	src := o.ComponentVersion
	dir := path.Dir(f)

	err = dest.PathFilesystem.MkdirAll(dir, 0o770)
	if err != nil {
		return err
	}

	format := formatoption.From(d.cmd)
	// FIXME: use CommonTransportFormat archives to store OCM components
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	set, err := comparch.Create(d.cmd.OCMContext(), accessobj.ACC_CREATE, f, format.Mode(), format.Format, accessio.PathFileSystem(dest.PathFilesystem))
	if err != nil {
		return err
	}
	//nolint:staticcheck // Deprecated: Component Archive (CA) - https://kubernetes.slack.com/archives/C05UWBE8R1D/p1734357630853489
	defer errors.PropagateError(&err, set.Close)

	nv := common.VersionedElementKey(src)
	hist := common.History{nv}

	err = transfer.CopyVersion(nil, d.cmd.OCMContext().Logger().WithValues("download", f), hist, src, set, nil)
	if err == nil {
		out.Outf(d.cmd.Context, "%s: downloaded\n", f)
	}
	return err
}

func DefaultFileName(obj *comphdlr.Object) string {
	f := obj.Spec.UniformRepositorySpec.String()
	f = strings.ReplaceAll(f, "::", "-")
	f = path.Join(f, obj.Spec.Component, *obj.Spec.Version)
	return f
}
