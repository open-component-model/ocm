package consumer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/signing"
)

const (
	Name = "consumer"
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <name>",
		Short: "determine consumer id",
		Long: `
Provide the required credential consumer id for signing context taken from 
*stdin*. The id is returned by *stdout*.
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
	Name string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	return nil
}

func Command(p ppi.Plugin, cmd *cobra.Command, opts *Options) error {
	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return err
	}

	var ctx ppi.SigningContext
	err = json.Unmarshal(data, &ctx)
	if err != nil {
		return err
	}

	h := p.GetSigningHandler(opts.Name)
	if h == nil {
		return errors.ErrUnknown(descriptor.KIND_SIGNING_HANDLER)
	}
	c := h.GetConsumerProvider()
	if c == nil {
		return fmt.Errorf("handler %q does not support consumer provider", opts.Name)
	}

	sctx := &signing.DefaultSigningContext{
		Hash:       ctx.HashAlgo,
		PrivateKey: ctx.PrivateKey,
		PublicKey:  ctx.PublicKey,
		Issuer:     ctx.Issuer,
	}

	cid := c(sctx)

	data, err = json.Marshal(cid)
	if err != nil {
		return err
	}
	cmd.OutOrStdout().Write(data)
	return nil
}
