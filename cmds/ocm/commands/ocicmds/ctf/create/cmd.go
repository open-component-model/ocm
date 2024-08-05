package create

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/cmds/ocm/commands/common/options/formatoption"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.TransportArchive
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	Format  formatoption.Option
	Handler ctf.FormatHandler
	Force   bool
	Path    string
}

// NewCommand creates a new ctf creation command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <path>",
		Args:  cobra.ExactArgs(1),
		Short: "create new OCI/OCM transport  archive",
		Long: `
Create a new empty OCM/OCI transport archive. This might be either a directory prepared
to host artifact content or a tar/tgz file.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.Format.AddFlags(fs)
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
}

func (o *Command) Complete(args []string) error {
	err := o.Format.Configure(o.Context)
	if err != nil {
		return err
	}
	o.Handler = ctf.GetFormat(o.Format.Format)
	if o.Handler == nil {
		return accessio.ErrInvalidFileFormat(o.Format.Format.String())
	}
	o.Path = args[0]
	return nil
}

func (o *Command) Run() error {
	mode := o.Format.Mode()
	fs := o.Context.FileSystem()
	if ok, err := vfs.Exists(fs, o.Path); ok || err != nil {
		if err != nil {
			return err
		}
		if o.Force {
			err = fs.RemoveAll(o.Path)
			if err != nil {
				return errors.Wrapf(err, "cannot remove old %q", o.Path)
			}
		}
	}
	obj, err := ctf.Create(o.Context.OCIContext(), accessobj.ACC_CREATE, o.Path, mode, o.Handler, fs)
	if err != nil {
		return err
	}
	if err != nil {
		obj.Close()
		return errors.Newf("creation failed: %s", err)
	}
	return obj.Close()
}
