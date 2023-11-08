// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package download

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/destoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	downloadcmd "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/extraid"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.CLI
	Verb  = verbs.Download
)

type Command struct {
	utils.BaseCommand

	ResourceType string

	Comp string
	Ids  []v1.Identity
	Path bool
}

// NewCommand creates a new CLI download command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	f := func(opts *output.Options) output.Output {
		return downloadcmd.NewAction(ctx, opts)
	}
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		versionconstraintsoption.New(true).SetLatest(),
		repooption.New(),
		output.OutputOptions(output.NewOutputs(f), downloadcmd.NewOptions(true).SetUseHandlers(), destoption.New()),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  [<component> {<name> { <key>=<value> }}]",
		Args:  cobra.MinimumNArgs(0),
		Short: "download OCM CLI from an OCM repository",
		Long: `
Download an OCM CLI executable. By default, the standard publishing component
and repository is used. Optionally, another component or repo and even a resource
can be specified. Resources are specified by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.

The option <code>-O</code> is used to declare the output destination.
The default location is the location of the <code>ocm</code> executable in
the actual PATH.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Path, "path", "p", false, "lookup executable in PATH")
}

func (o *Command) Complete(args []string) error {
	var err error

	if len(args) > 0 {
		o.Comp = args[0]
	} else {
		o.Comp = COMPONENT
	}
	if len(args) > 1 {
		o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	}
	if err == nil {
		if o.ResourceType == "" {
			o.ResourceType = resourcetypes.EXECUTABLE
		}
		if len(o.Ids) == 0 {
			o.Ids = []v1.Identity{
				{
					v1.SystemIdentityName: RESOURCE,
				},
			}
		}
		for _, id := range o.Ids {
			id[extraid.ExecutableOperatingSystem] = runtime.GOOS
			id[extraid.ExecutableArchitecture] = runtime.GOARCH
		}
	}

	return err
}

func (o *Command) handlerOptions() []elemhdlr.Option {
	return []elemhdlr.Option{common.WithTypes([]string{o.ResourceType})}
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	dest := destoption.From(o)
	if dest.Destination == "" {
		p := os.Getenv("OCMCMD")
		if p == "" && !o.Path {
			p, err = os.Executable()
			if err != nil {
				out.Outln(o, "WARNING: cannot detect actual executable (%w) -> fallback to PATH lookup")
			}
		}
		if p == "" {
			list := utils.SplitPathList(os.ExpandEnv(os.Getenv("PATH")))
			for _, e := range list {
				t := filepath.Join(e, "ocm"+EXECUTABLE_SUFFIX)
				if utils.IsExecutable(t, o.FileSystem()) {
					p = t
					break
				}
			}
		}
		if p == "" {
			return fmt.Errorf("no download target for OCM CLI command found")
		} else {
			out.Outln(o, "updating OCM CLI command at", p)
			dest.Destination = p
		}
	} else {
		if ok, err := vfs.IsDir(o.FileSystem(), dest.Destination); ok && err == nil {
			dest.Destination = vfs.Join(o.FileSystem(), dest.Destination, "ocm"+EXECUTABLE_SUFFIX)
		}
	}
	opts := output.From(o)

	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp}, o.handlerOptions()...)
	if err != nil {
		return err
	}
	specs, err := utils.ElemSpecs(o.Ids)
	if err != nil {
		return err
	}

	return utils.HandleOutputs(opts, hdlr, specs...)
}
