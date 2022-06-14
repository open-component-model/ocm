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

package transfer

import (
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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
		formatoption.New(),
		overwriteoption.New(),
		rscbyvalueoption.New(),
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

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o, session))
	if err != nil {
		return err
	}

	src, err := ctf.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.SourceName, 0, o.FileSystem())
	if err != nil {
		return errors.Wrapf(err, "cannot open source")
	}
	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, formatoption.From(o).Format, o.Context.FileSystem())
	if err != nil {
		return err
	}

	thdlr, err := spiff.New(
		spiff.Script(scriptoption.From(o).ScriptData),
		rscbyvalueoption.From(o),
		overwriteoption.From(o),
		spiff.ScriptFilesystem(o.FileSystem()),
	)
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
	cmd     *Command
	printer common.Printer
	target  ocm.Repository
	handler transferhandler.TransferHandler
	closure transfer.TransportClosure
	errors  *errors.ErrorList
}

func (a *action) Execute(src ocm.Repository) error {
	return transfer.TransferComponents(a.printer, a.closure, src, "", true, a.target, a.handler)
}
