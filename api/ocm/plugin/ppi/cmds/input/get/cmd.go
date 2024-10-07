package get

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	commonppi "ocm.software/ocm/api/ocm/plugin/ppi/cmds/common"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	Name     = "get"
	OptCreds = commonppi.OptCreds
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " [<flags>] [<dir>] <access spec>",
		Short: "get blob",
		Long: `
Evaluate the given input specification and return the described blob on
*stdout*.`,
		Args: cobra.RangeArgs(1, 2),
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
	Dir           string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	flag.YAMLVarP(fs, &o.Credentials, OptCreds, "c", nil, "credentials")
	flag.StringToStringVarPFA(fs, &o.Credentials, "credential", "C", nil, "dedicated credential value")
}

func (o *Options) Complete(args []string) error {
	creds := 0
	if len(args) > 1 {
		o.Dir = args[creds]
		creds++
	} else {
		o.Dir = "."
	}
	if err := runtime.DefaultYAMLEncoding.Unmarshal([]byte(args[creds]), &o.Specification); err != nil {
		return errors.Wrapf(err, "invalid repository specification")
	}

	fmt.Fprintf(os.Stderr, "credentials: %s\n", o.Credentials.String())
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	spec, err := p.DecodeInputSpecification(opts.Specification)
	if err != nil {
		return errors.Wrapf(err, "access specification")
	}

	m := p.GetInputType(spec.GetType())
	if m == nil {
		return errors.ErrUnknown(descriptor.KIND_INPUTTYPE, spec.GetType())
	}
	_, err = m.ValidateSpecification(p, opts.Dir, spec)
	if err != nil {
		return err
	}
	r, err := m.Reader(p, opts.Dir, spec, opts.Credentials)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, r)
	r.Close()
	return err
}
