// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/formatoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/overwriteoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/rscbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/skipupdateoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/srcbyvalueoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
)

var (
	Names = names.ComponentArchive
	Verb  = verbs.Transfer
)

type Command struct {
	utils.BaseCommand
	Path       string
	TargetName string
}

// NewCommand creates a new transfer command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{BaseCommand: utils.NewBaseCommand(ctx,
			formatoption.New(),
			lookupoption.New(),
			skipupdateoption.New(),
			overwriteoption.New(),
			rscbyvalueoption.New(),
			srcbyvalueoption.New(),
		)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <source> <target>",
		Args:  cobra.MinimumNArgs(2),
		Short: "transfer component archive to some component repository",
		Long: `
Transfer a component archive to some component repository. This might
be a CTF Archive or a regular repository.
If the type CTF is specified the target must already exist, if CTF flavor
is specified it will be created if it does not exist.

Besides those explicitly known types a complete repository spec might be configured,
either via inline argument or command configuration file and name.
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.Path = args[0]
	o.TargetName = args[1]

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

	source, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.Path, 0, o.Context)
	if err != nil {
		return err
	}
	session.Closer(source)

	format := formatoption.From(o).ChangedFormat()
	target, err := ocm.AssureTargetRepository(session, o.Context.OCMContext(), o.TargetName, ocm.CommonTransportFormat, format, o.Context.FileSystem())
	if err != nil {
		return err
	}

	transferopts := &standard.Options{}
	transferhandler.From(o.ConfigContext(), transferopts)
	transferhandler.ApplyOptions(transferopts, options.FindOptions[transferhandler.TransferOption](o)...)
	thdlr, err := standard.New(transferopts)
	if err != nil {
		return err
	}
	return transfer.TransferVersion(common.NewPrinter(o.Context.StdOut()), nil, source, target, thdlr)
}
