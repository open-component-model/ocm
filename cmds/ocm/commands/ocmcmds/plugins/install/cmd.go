package install

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/mandelsoft/goutils/errors"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/cobrautils/flag"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Plugins
	Verb  = verbs.Install
)

type Command struct {
	utils.BaseCommand

	cache.PluginUpdater

	Ref   string
	Names []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx, repooption.New()),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component version ref> [<name>] | <name>",
		Short: "install or update an OCM plugin",
		Long: `
Download and install a plugin provided by an OCM component version.
For the update mode only the plugin name is required. 

If no version is specified the latest version is chosen. If at least one
version constraint is given, only the matching versions are considered.
`,
		Args: cobra.MinimumNArgs(1),
		Example: `
$ ocm install plugin ghcr.io/github.com/mandelsoft/cnudie//github.com/mandelsoft/ocmplugin:0.1.0-dev
$ ocm install plugin -c 1.2.x ghcr.io/github.com/mandelsoft/cnudie//github.com/mandelsoft/ocmplugin
$ ocm install plugin -u demo
$ ocm install plugin -r demo
`,
	}
}
func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.RemoveMode, "remove", "r", false, "remove plugin")
	fs.BoolVarP(&o.UpdateMode, "update", "u", false, "update plugin")
	fs.BoolVarP(&o.Describe, "describe", "d", false, "describe plugin, only")
	fs.BoolVarP(&o.Force, "force", "f", false, "overwrite existing plugin")
	flag.SemverConstraintsVarP(fs, &o.Constraints, "constraints", "c", nil, "version constraint")
}

func (o *Command) Complete(args []string) error {
	o.PluginUpdater.Context = o.BaseCommand.OCMContext()

	if o.UpdateMode && o.RemoveMode {
		return fmt.Errorf("either remove or update mode possible, only")
	}
	if o.UpdateMode || o.RemoveMode {
		o.Names = args
		if len(args) == 0 {
			return fmt.Errorf("for update mode the plugin name is required")
		}
	} else {
		if len(args) > 2 {
			return fmt.Errorf("only two arguments (<component version> [<resource name>]) possible")
		}
		if len(args) > 1 {
			o.Names = []string{args[1]}
		}
		o.Ref = args[0]
	}
	o.Printer = common.NewPrinter(o.StdOut())
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o.BaseCommand.Context, session))
	if err != nil {
		return err
	}

	repo := repooption.From(o)

	if o.UpdateMode || o.RemoveMode {
		msg := []string{"updating", "updated"}
		f := o.Update
		if o.RemoveMode {
			msg = []string{"removing", "removed"}
			f = o.Remove
		}
		pi := plugincacheattr.Get(o.OCMContext())
		failed := 0
		for _, n := range o.Names {
			o.Ref = n
			if pi.Get(n) == nil {
				objs := pluginhdlr.Lookup(n, pi)
				if len(objs) == 1 {
					o.Ref = pluginhdlr.Elem(objs[0]).Name()
				}
			}
			if err := f(session, o.Ref); err != nil {
				if len(o.Names) == 1 {
					return errors.Wrapf(err, "%s plugin %s failed", msg[0], o.Ref)
				}
				out.Errf(o, "%s plugin %s failed: %s\n", msg[0], o.Ref, err.Error())
				failed++
			}
		}
		if failed > 0 {
			return fmt.Errorf("%s of %d plugin(s) failed", msg[0], failed)
		}
		return nil
	}
	name := ""
	if len(o.Names) > 0 {
		name = o.Names[0]
	}
	if repo.Repository != nil {
		return o.DownloadFromRepo(session, repo.Repository, o.Ref, name)
	}
	return o.DownloadRef(session, o.Ref, name)
}

/////////////////////////////////////////////////////////////////////////////
