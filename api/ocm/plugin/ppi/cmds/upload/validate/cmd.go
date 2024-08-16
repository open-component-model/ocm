package validate

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name     = "validate"
	OptMedia = common.OptMedia
	OptArt   = common.OptArt
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <name> <spec>",
		Short: "validate upload specification",
		Long: `
This command accepts a target specification as argument. It is used to
validate the specification for the specified uploader and to provide some
metadata for the given specification.

This metadata has to be provided as JSON document string on *stdout* and has the
following fields:

- **<code>consumerId</code>** *map[string]string*

  The consumer id used to determine optional credentials for the
  underlying repository. If specified, at least the <code>type</code> field must
  be set.
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
	Name          string
	Specification json.RawMessage

	ArtifactType string
	MediaType    string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}
	return nil
}

type Result struct {
	ConsumerId credentials.ConsumerIdentity `json:"consumerId"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeUploadTargetSpecification(opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "target specification")
	}

	m := p.GetUploader(opts.Name)
	if m == nil {
		return errors.ErrUnknown(descriptor.KIND_UPLOADER, spec.GetType())
	}
	info, err := m.ValidateSpecification(p, spec)
	if err != nil {
		return err
	}
	result := Result{info.ConsumerId}
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
