package validate

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/runtime"
)

const Name = "validate"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <spec>",
		Short: "validate value set",
		Long: `
This command accepts a value set as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the 
following fields: 

- **<code>description</code>** *string*

  A short textual description of the described value set.
`,
		Args: cobra.ExactArgs(2),
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
	Purpose       string
	Specification json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Purpose = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid valueset specification")
	}
	return nil
}

type Result struct {
	Short string `json:"description"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeValueSet(opts.Purpose, opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "access specification")
	}

	k, v := runtime.KindVersion(spec.GetType())
	m := p.GetValueSet(opts.Purpose, k, v)
	if m == nil {
		return errors.ErrUnknown(descriptor.KIND_VALUESET, spec.GetType())
	}
	info, err := m.ValidateSpecification(p, spec)
	if err != nil {
		return err
	}
	result := Result{Short: info.Short}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
