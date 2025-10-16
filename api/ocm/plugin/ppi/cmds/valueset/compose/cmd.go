package compose

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/runtime"
)

const Name = "compose"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <purpose> <name> <options json> <base spec json>",
		Short: "compose value set from options and base specification",
		Long: `
The task of this command is used to compose and validate a value set based on
some explicitly given input options and preconfigured specifications.

The finally composed set has to be returned as JSON document
on *stdout*.

This command is only used, if for a value set descriptor configuration
na direct composition rules are configured (<CMD>` + p.Name() + ` descriptor</CMD>).

If possible, predefined standard options should be used. In such a case only the
<code>name</code> field should be defined for an option. If required, new options can be
defined by additionally specifying a type and a description. New options should
be used very carefully. The chosen names MUST not conflict with names provided
by other plugins. Therefore, it is highly recommended to use names prefixed
by the plugin name.

` + options.DefaultRegistry.Usage(),
		Args: cobra.ExactArgs(4),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Command(p, cmd, &opts)
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

type Options struct {
	Purpose string
	Name    string
	Options ppi.Config
	Base    ppi.Config
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Purpose = args[0]
	o.Name = args[1]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[2]), &o.Options); err != nil {
		return errors.Wrapf(err, "invalid avalue set options")
	}
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[3]), &o.Base); err != nil {
		return errors.Wrapf(err, "invalid base set specification")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	k, v := runtime.KindVersion(opts.Name)
	s := p.GetValueSet(opts.Purpose, k, v)
	if s == nil {
		return errors.ErrUnknown(descriptor.KIND_VALUESET, opts.Name)
	}
	err := opts.Options.ConvertFor(s.Options()...)
	if err != nil {
		return err
	}
	err = s.ComposeSpecification(p, opts.Options, opts.Base)
	if err != nil {
		return err
	}
	data, err := json.Marshal(opts.Base)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
