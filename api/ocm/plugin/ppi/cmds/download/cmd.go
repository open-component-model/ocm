package download

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
)

const (
	Name      = "download"
	OptMedia  = common.OptMedia
	OptArt    = common.OptArt
	OptConfig = common.OptConfig
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] <name> <filepath>",
		Short: "download blob into filesystem",
		Long: `
This command accepts a target filepath as argument. It is used as base name
to store the downloaded content. The blob content is provided on the
*stdin*. The first argument specified the downloader to use for the operation.

The task of this command is to transform the content of the provided 
blob into a filesystem structure applicable to the type specific tools working
with content of the given artifact type.
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
	Name string
	Path string

	MediaType    string
	ArtifactType string
	Config       string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.MediaType, OptMedia, "m", "", "media type of input blob")
	fs.StringVarP(&o.ArtifactType, OptArt, "a", "", "artifact type of input blob")
	fs.StringVarP(&o.Config, OptConfig, "c", "", "registration config")
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	o.Path = args[1]
	return nil
}

type Result struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	d := p.GetDownloader(opts.Name)
	if d == nil {
		return errors.ErrNotFound(descriptor.KIND_DOWNLOADER, fmt.Sprintf("%s:%s", opts.ArtifactType, opts.MediaType))
	}
	var cfg []byte
	if opts.Config != "" {
		cfg = []byte(opts.Config)
	}
	w, h, err := d.Writer(p, opts.ArtifactType, opts.MediaType, opts.Path, cfg)
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
	path, err := h()
	result := Result{
		Path: path,
	}
	if err != nil {
		result.Error = err.Error()
	}
	data, err := json.Marshal(result)
	if err == nil {
		cmd.Printf("%s\n", string(data))
	}
	return err
}
