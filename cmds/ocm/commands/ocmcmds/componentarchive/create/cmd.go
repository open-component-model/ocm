// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/errors"
)

var (
	Names = names.ComponentArchive
	Verb  = verbs.Create
)

type Command struct {
	utils.BaseCommand

	providerattrs []string

	Handler comparch.FormatHandler
	Force   bool
	Path    string
	Format  string

	Component      string
	Version        string
	Provider       string
	ProviderLabels metav1.Labels
	Labels         metav1.Labels
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, formatoption.New(comparch.GetFormats()...), schemaoption.New(compdesc.DefaultSchemeVersion))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component> <version> --provider <provider-name> {--provider <label>=<value>} {<label>=<value>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "create new component archive",
		Long: `
Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file (see option --type).

A provider must be specified, additional provider labels are optional.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
	fs.StringArrayVarP(&o.providerattrs, "provider", "p", nil, "provider attribute")
	fs.StringVarP(&o.Path, "file", "F", "component-archive", "target file/directory")
}

func (o *Command) Complete(args []string) error {
	var err error

	format := formatoption.From(o).Format
	o.Handler = comparch.GetFormat(format)
	if o.Handler == nil {
		return accessio.ErrInvalidFileFormat(format.String())
	}

	o.Component = args[0]
	o.Version = args[1]

	for _, a := range args[2:] {
		o.Labels, err = common.AddParsedLabel(o.Labels, a)
		if err != nil {
			return err
		}
	}
	for _, a := range o.providerattrs {
		if !strings.Contains(a, "=") {
			if o.Provider != "" {
				return fmt.Errorf("%s: provider name is already set (%s)", a, o.Provider)
			}
			o.Provider = a
			continue
		}
		o.ProviderLabels, err = common.AddParsedLabel(o.ProviderLabels, a)
		if err != nil {
			return err
		}
	}
	if o.Provider == "" {
		return fmt.Errorf("provider name missing")
	}
	return nil
}

func (o *Command) Run() error {
	mode := formatoption.From(o).Mode()
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
	obj, err := comparch.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, o.Path, mode, o.Handler, fs)
	if err != nil {
		fs.RemoveAll(o.Path)
		return err
	}
	desc := obj.GetDescriptor()
	desc.Metadata.ConfiguredVersion = schemaoption.From(o).Schema
	desc.Name = o.Component
	desc.Version = o.Version
	desc.Provider.Name = metav1.ProviderName(o.Provider)
	desc.Provider.Labels = o.ProviderLabels
	desc.Labels = o.Labels

	err = compdesc.Validate(desc)
	if err != nil {
		obj.Close()
		fs.RemoveAll(o.Path)
		return errors.Newf("invalid component info: %s", err)
	}
	err = obj.Close()
	if err != nil {
		fs.RemoveAll(o.Path)
	}
	return err
}
