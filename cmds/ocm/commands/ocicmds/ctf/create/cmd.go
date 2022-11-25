// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
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
	err := o.Format.Complete(o.Context)
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
