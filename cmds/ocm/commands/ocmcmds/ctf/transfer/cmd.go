// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/stoponexistingoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/errors"
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
		lookupoption.New(),
		formatoption.New(),
		overwriteoption.New(),
		rscbyvalueoption.New(),
		stoponexistingoption.New(),
		uploaderoption.New(),
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
		spiff.Script(scriptoption.From(o).ScriptData),
		closureoption.From(o),
		lookupoption.From(o),
		rscbyvalueoption.From(o),
		stoponexistingoption.From(o),
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
	printer common.Printer
	target  ocm.Repository
	handler transferhandler.TransferHandler
	closure transfer.TransportClosure
	errors  *errors.ErrorList
}

func (a *action) Execute(src ocm.Repository) error {
	return transfer.TransferComponents(a.printer, a.closure, src, "", true, a.target, a.handler)
}
