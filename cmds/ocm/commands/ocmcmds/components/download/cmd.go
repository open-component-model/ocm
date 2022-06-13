// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package download

import (
	"path"
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
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
	o.Refs = args[1:]
	return err
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o.Context, session))
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
	if len(d.data) == 1 {
		return d.Save(d.data[0], dest.Destination)
	} else {
		for _, e := range d.data {
			f := e.Spec.UniformRepositorySpec.String()
			f = strings.ReplaceAll(f, "::", "-")
			f = path.Join(f, e.Spec.Component, *e.Spec.Version)
			err := d.Save(e, f)
			if err != nil {
				list.Add(err)
				out.Outf(d.cmd.Context, "%s failed: %s\n", f, err)
			}
		}
	}
	return list.Result()
}

func (d *action) Save(o *comphdlr.Object, f string) error {
	dest := destoption.From(d.cmd)
	src := o.ComponentVersion
	dir := path.Dir(f)

	err := dest.PathFilesystem.MkdirAll(dir, 0770)
	if err != nil {
		return err
	}

	format := formatoption.From(d.cmd)
	set, err := comparch.Create(d.cmd.OCMContext(), accessobj.ACC_CREATE, f, format.Mode(), format.Format, accessio.PathFileSystem(dest.PathFilesystem))
	if err != nil {
		return err
	}
	defer set.Close()

	nv := common.NewNameVersion(src.GetName(), src.GetVersion())
	hist := common.History{nv}

	err = transfer.CopyVersion(nil, hist, src, set, nil)
	if err == nil {
		out.Outf(d.cmd.Context, "%s: downloaded\n", f)
	}
	return err
}
