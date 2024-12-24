package put

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/refhints"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name     = "put"
	OptCreds = common.OptCreds
	OptHint  = common.OptHint
	OptMedia = common.OptMedia
	OptArt   = common.OptArt
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

	hint string

	Hints ppi.ReferenceHints
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, OptCreds, "c", nil, "credentials")
	flag.StringToStringVarPF(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
	fs.StringVarP(&o.hint, OptHint, "H", "", "reference hint for storing blob")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[1]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}
	o.Hints = refhints.ParseHints(o.hint)
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeUploadTargetSpecification(opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "target specification")
	}

	u := p.GetUploader(opts.Name)
	if u == nil {
		return errors.ErrNotFound(descriptor.KIND_UPLOADER, fmt.Sprintf("%s:%s", opts.ArtifactType, opts.MediaType))
	}
	w, h, err := u.Writer(p, opts.ArtifactType, opts.MediaType, opts.Hints, spec, opts.Credentials)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, os.Stdin)
	if err != nil {
		w.Close()
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	acc := h()
	data, err := json.Marshal(acc)
	if err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
