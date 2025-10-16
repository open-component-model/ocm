package install

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin/cache"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
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
