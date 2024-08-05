package validate

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/utils/errkind"
	"ocm.software/ocm/api/utils/runtime"
)

const Name = "validate"

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <spec>",
		Short: "validate access specification",
		Long: `
This command accepts an access specification as argument. It is used to
validate the specification and to provide some metadata for the given
specification.

This metadata has to be provided as JSON string on *stdout* and has the 
following fields: 

- **<code>mediaType</code>** *string*

  The media type of the artifact described by the specification. It may be part
  of the specification or implicitly determined by the access method.

- **<code>description</code>** *string*

  A short textual description of the described location.

- **<code>hint</code>** *string*

  A name hint of the described location used to reconstruct a useful
  name for local blobs uploaded to a dedicated repository technology.

- **<code>consumerId</code>** *map[string]string*

  The consumer id used to determine optional credentials for the
  underlying repository. If specified, at least the <code>type</code> field must be set.
`,
		Args: cobra.ExactArgs(1),
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
	Specification json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[0]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid access specification")
	}
	return nil
}

type Result struct {
	MediaType  string                       `json:"mediaType"`
	Short      string                       `json:"description"`
	Hint       string                       `json:"hint"`
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeAccessSpecification(opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "access specification")
	}

	m := p.GetAccessMethod(runtime.KindVersion(spec.GetType()))
	if m == nil {
		return errors.ErrUnknown(errkind.KIND_ACCESSMETHOD, spec.GetType())
	}
	info, err := m.ValidateSpecification(p, spec)
	if err != nil {
		return err
	}
	result := Result{MediaType: info.MediaType, ConsumerId: info.ConsumerId, Hint: info.Hint, Short: info.Short}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
