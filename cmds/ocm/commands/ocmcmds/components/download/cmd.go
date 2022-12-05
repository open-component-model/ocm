// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"path"
	"strings"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
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
	if len(d.data) == 1 {
		f := dest.Destination
		if f == "" {
			f = DefaultFileName(d.data[0]) + format.Format.Suffix()
		}
		return d.Save(d.data[0], f)
	} else {
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
	set, err := comparch.Create(d.cmd.OCMContext(), accessobj.ACC_CREATE, f, format.Mode(), format.Format, accessio.PathFileSystem(dest.PathFilesystem))
	if err != nil {
		return err
	}
	defer errors.PropagateError(&err, set.Close)

	nv := common.NewNameVersion(src.GetName(), src.GetVersion())
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
