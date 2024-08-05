package execute

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/valuemergehandler/hpi"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name = "execute"
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <name> <spec>",
		Short: "execute a value merge",
		Long: `
This command executes a value merge. The values are taken from *stdin* as JSON
string. It has the following fields: 

- **<code>local</code>** *any*

  The local value to merge into the inbound value.

- **<code>inbound</code>** *any*

  The value to merge into. This value is based on the original inbound value.

This action has to provide an execution result as JSON string on *stdout*. It has the 
following fields: 

- **<code>modified</code>** *bool*

  Whether the inbound value has been modified by merging with the local value.

- **<code>value</code>** *string*

  The merged value

- **<code>message</code>** *string*

  An error message.
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
	Configuration json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("algorithm name missing")
	}
	o.Name = args[0]
	if len(args) > 1 {
		o.Configuration = []byte(args[1])
	}
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	h := p.GetValueMergeHandler(opts.Name)
	if h == nil {
		return errors.ErrUnknown(hpi.KIND_VALUE_MERGE_ALGORITHM, opts.Name)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	var input ppi.ValueMergeData
	err = json.Unmarshal(data, &input)
	if err != nil {
		return err
	}

	result, err := h.Execute(p, input.Local, input.Inbound, opts.Configuration)
	if err != nil {
		return err
	}
	data, err = runtime.DefaultJSONEncoding.Marshal(result)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
