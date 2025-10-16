package utils

import (
	"reflect"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/general"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/utils/cobrautils"
	"ocm.software/ocm/cmds/ocm/common/options"
)

// OCMCommand is a command pattern, that can be instantiated for a dediated
// sub command name to create a cobra command.
type OCMCommand interface {
	clictx.Context

	// ForName create a new cobra command for the given command name.
	// The Use attribute should omit the command name and just provide
	// a ost argument synopsis.
	// the complete attribute set is tweaked with the SetupCommand function
	// which calls this method.
	// Basically this should be an inherited function by the base implementation
	// but GO does not support virtual methods, therefore it is a global
	// function instead of a method.
	ForName(name string) *cobra.Command
	AddFlags(fs *pflag.FlagSet)
	Complete(args []string) error
	Run() error
}

////////////////////////////////////////////////////////////////////////////////
// optional interfaces for a Command

// Updater is a generic update function.
type Updater interface {
	Update(cmd *cobra.Command, args []string)
}

// Long is used to provide the long description by a method of the
// command.
type Long interface {
	Long() string
}

// Short is used to provide the short description by a method of the
// command.
type Short interface {
	Short() string
}

// Example is used to provide the example description by a method of the
// command.
type Example interface {
	Example() string
}

type CommandTweaker interface {
	TweakCommand(cmd *cobra.Command)
}

////////////////////////////////////////////////////////////////////////////////

// BaseCommand provides the basic functionality of an OCM command
// to carry a context and a set of reusable option specs.
type BaseCommand struct {
	clictx.Context
	options.OptionSet
}

func NewBaseCommand(ctx clictx.Context, opts ...options.Options) BaseCommand {
	return BaseCommand{Context: ctx, OptionSet: opts}
}

func (BaseCommand) Complete(args []string) error { return nil }

func addCommand(names []string, use string) string {
	if use == "" {
		return names[0]
	}
	lines := strings.Split(use, "\n")
outer:
	for i, l := range lines {
		if strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			continue
		}
		for _, n := range names {
			if strings.HasPrefix(l, n+" ") {
				continue outer
			}
		}
		lines[i] = names[0] + " " + l
	}
	return strings.Join(lines, "\n")
}

func HideCommand(cmd *cobra.Command) *cobra.Command {
	cmd.Hidden = true
	return cmd
}

func OverviewCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["overview"] = ""
	return cmd
}

func DocuCommandPath(cmd *cobra.Command, path string, optname ...string) *cobra.Command {
	name := general.OptionalDefaulted(cmd.Name(), optname...)
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations["commandPath"] = path + " " + name
	return cmd
}

func MassageCommand(cmd *cobra.Command, names ...string) *cobra.Command {
	cmd.Use = addCommand(names, cmd.Use)
	if len(names) > 1 {
		cmd.Aliases = names[1:]
	}
	cmd.DisableFlagsInUseLine = true
	cmd.TraverseChildren = true
	return cmd
}

// updateFunction composes an update function for field setter or a generic
// update function used to update cobra command attributes based on optional
// methods provided by the OCM command.
// The generated update function is then called to update attributes of the
// cobra command prior to help output.
// If field setters are found, their values will override explicit settings
// configured for the cobra command structure.
// If no updaters are found, the statically configured values at the cobra
// command struct will be used.
func updateFunction[T any](f func(cmd *cobra.Command, args []string), source OCMCommand, target *cobra.Command) func(cmd *cobra.Command, args []string) {
	var rf func(cmd *cobra.Command, args []string)

	if t, ok := source.(T); ok {
		var i *T
		name := reflect.TypeOf(i).Elem().Method(0).Name
		m := reflect.ValueOf(t).MethodByName(name)

		tv := reflect.ValueOf(target).Elem()
		if field := tv.FieldByName(name); field.IsValid() {
			rf = func(cmd *cobra.Command, args []string) {
				field.Set(m.Call([]reflect.Value{})[0])
			}
			rf(target, nil)
		} else {
			rf = func(cmd *cobra.Command, args []string) {
				m.Call([]reflect.Value{reflect.ValueOf(cmd), reflect.ValueOf(args)})
			}
		}
	}
	if f == nil {
		return rf
	}
	if rf == nil {
		return f
	}
	return func(cmd *cobra.Command, args []string) {
		f(cmd, args)
		rf(cmd, args)
	}
}

// SetupCommand uses the OCMCommand to create and tweak a cobra command
// to incorporate the additional reusable option specs and their usage documentation.
// Before the command executions the various Complete method flavors are
// executed on the additional options on the OCMCommand.
// It also prepares the help system to reflect dynamic settings provided
// by root command options by using a generated update function based
// on optional methods of the OCM command.
func SetupCommand(ocmcmd OCMCommand, names ...string) *cobra.Command {
	c := ocmcmd.ForName(names[0])
	MassageCommand(c, names...)

	var update func(cmd *cobra.Command, args []string)

	update = updateFunction[Long](update, ocmcmd, c)
	update = updateFunction[Short](update, ocmcmd, c)
	update = updateFunction[Example](update, ocmcmd, c)
	update = updateFunction[Updater](update, ocmcmd, c)

	c.RunE = func(cmd *cobra.Command, args []string) error {
		var err error
		if set, ok := ocmcmd.(options.OptionSetProvider); ok {
			err = set.AsOptionSet().ProcessOnOptions(options.CompleteOptionsWithCLIContext(ocmcmd))
		}
		if err == nil {
			err = ocmcmd.Complete(args)
			if err == nil {
				err = ocmcmd.Run()
			}
		}
		/*
			if err != nil && ocmcmd.StdErr() != os.Stderr {
				out.Error(ocmcmd, err.Error())
			}
		*/
		list := errors.ErrListf("")
		list.Add(err)
		list.Add(ocmcmd.OCMContext().Finalize())
		return list.Result()
	}
	if t, ok := ocmcmd.(CommandTweaker); ok {
		t.TweakCommand(c)
	}
	if u, ok := ocmcmd.(options.Usage); ok {
		c.Long += u.Usage()
	}

	cobrautils.CleanMarkdownUsageFunc(c)
	help := c.HelpFunc()
	if update != nil {
		c.SetHelpFunc(func(cmd *cobra.Command, args []string) {
			// assure root options are evaluated to provide appropriate base
			// for update functions. PreRun functions will not be called
			// if option --help is used to call the help function, so just
			// call it in such a case.
			for _, a := range args {
				if a == "--help" {
					root := cmd.Root()
					if cmd != root {
						if root.PersistentPreRunE != nil {
							root.PersistentPreRunE(cmd, args)
						} else if root.PersistentPreRun != nil {
							root.PersistentPreRun(cmd, args)
						}
					}
					break
				}
			}
			update(cmd, args)
			help(cmd, args)
		})
	}
	ocmcmd.AddFlags(c.Flags())
	return c
}

func Names(def []string, names ...string) []string {
	if len(names) == 0 {
		return def
	}
	return names
}
