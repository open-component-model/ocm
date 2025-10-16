package put

import (
	"encoding/json"
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name      = "put"
	OptCreds  = common.OptCreds
	OptHint   = common.OptHint
	OptMedia  = common.OptMedia
	OptArt    = common.OptArt
	OptDigest = common.OptDigest
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <name> <repository specification>",
		Short: "upload blob to external repository",
		Long: `
Read the blob content from *stdin*, store the blob in the repository specified
by the given repository specification and return the access specification
(as JSON document string) usable to retrieve the blob, again, on * stdout*.
The uploader to use is specified by the first argument. This might only be
relevant, if the plugin supports multiple uploaders.
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

	Credentials  credentials.DirectCredentials
	MediaType    string
	ArtifactType string
	Digest       string

	Hint string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, OptCreds, "c", nil, "credentials")
	flag.StringToStringVarPF(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
	fs.StringVarP(&o.Hint, OptHint, "H", "", "reference hint for storing blob")
	fs.StringVarP(&o.Digest, OptDigest, "d", "", "digest of the blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return fmt.Errorf("invalid repository specification: %w", err)
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) (err error) {
	var spec ppi.UploadTargetSpec
	if spec, err = p.DecodeUploadTargetSpecification(opts.Specification); err != nil {
		return fmt.Errorf("error decoding upload target specification: %w", err)
	}

	u := p.GetUploader(opts.Name)
	if u == nil {
		return errors.ErrNotFound(descriptor.KIND_UPLOADER, fmt.Sprintf("%s:%s", opts.ArtifactType, opts.MediaType))
	}

	reader := cmd.InOrStdin()

	var provider ppi.AccessSpecProvider
	if provider, err = u.Upload(
		cmd.Context(),
		p,
		opts.ArtifactType,
		opts.MediaType,
		opts.Hint,
		opts.Digest,
		spec,
		opts.Credentials,
		reader,
	); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	acc := provider()

	var data []byte
	if data, err = json.Marshal(acc); err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
