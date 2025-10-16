package execute

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/action"
	"ocm.software/ocm/api/datacontext/action/api"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name     = "execute"
	OptCreds = common.OptCreds
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <spec>",
		Short: "execute an action",
		Long: `
This command executes an action.

This action has to provide an execution result as JSON string on *stdout*. It has the 
following fields: 

- **<code>name</code>** *string*

  The name and version of the action result. It must match the value
  from the action specification.

- **<code>message</code>** *string*

  An error message.

Additional fields depend on the kind of action.
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
	Credentials   credentials.DirectCredentials
	Specification json.RawMessage
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, OptCreds, "c", nil, "credentials")
	flag.StringToStringVarPFA(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
}

func (o *Options) Complete(args []string) error {
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[0]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid access specification")
	}
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := action.DefaultRegistry().DecodeActionSpec(opts.Specification, runtime.DefaultJSONEncoding)
	if err != nil {
		return errors.Wrapf(err, "action specification")
	}

	a := p.GetAction(spec.GetKind())
	if a == nil {
		return errors.ErrUnknown(api.KIND_ACTION, spec.GetKind())
	}
	result, err := a.Execute(p, spec, opts.Credentials)
	if err != nil {
		return err
	}
	result.SetType(spec.GetType())
	data, err := action.DefaultRegistry().EncodeActionResult(result, runtime.DefaultJSONEncoding)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", string(data))
	return nil
}
