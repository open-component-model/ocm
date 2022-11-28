// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/scriptoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/uploaderoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/spiff"
	"github.com/open-component-model/ocm/pkg/errors"
)

var (
	Names = names.Components
	Verb  = verbs.Transfer
)

type Command struct {
	utils.BaseCommand

	Refs       []string
	TargetName string
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
		rscbyvalueoption.New(),
		uploaderoption.New(),
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
$ ocm transfer components -t tgz --repo OCIRegistry:ghcr.io mandelsoft/kubelink ctf.tgz
`,
	}
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
	
	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	err = uploaderoption.From(o).Register(o.Context)
	if err != nil {
		return err
	}

	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, formatoption.From(o).Format, o.Context.FileSystem())
	if err != nil {
		return err
	}

	thdlr, err := spiff.New(
		closureoption.From(o),
		overwriteoption.From(o),
		rscbyvalueoption.From(o),
		lookupoption.From(o),
		spiff.Script(scriptoption.From(o).ScriptData),
		spiff.ScriptFilesystem(o.FileSystem()),
	)
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
	o := e.(*comphdlr.Object)
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
	return nil
}
