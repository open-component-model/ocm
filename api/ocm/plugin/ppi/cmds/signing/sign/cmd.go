package sign

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	"ocm.software/ocm/api/ocm/plugin/ppi"
	"ocm.software/ocm/api/tech/signing"
)

const (
	Name = "sign"
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <name> <digest>",
		Short: "sign digest",
		Long: `
Sign the given digest with signing context taken from *stdin* and return the signature on *stdout*.`,
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
	Name   string
	Digest string
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	o.Name = args[0]
	o.Digest = args[1]
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
	s := h.GetSigner()
	if s == nil {
		return fmt.Errorf("handler %q does not support signing", opts.Name)
	}

	sctx := &signing.DefaultSigningContext{
		Hash:       ctx.HashAlgo,
		PrivateKey: ctx.PrivateKey,
		PublicKey:  ctx.PublicKey,
		Issuer:     ctx.Issuer,
	}

	cctx := credentials.New(datacontext.MODE_EXTENDED)
	if h.GetConsumerProvider() != nil && ctx.Credentials != nil {
		cid := h.GetConsumerProvider()(sctx)
		cctx.SetCredentialsForConsumer(cid, ctx.Credentials)
	}

	sig, err := s.Sign(cctx, opts.Digest, sctx)
	if err != nil {
		return err
	}

	spec := ppi.SignatureSpecFor(sig)
	out, err := json.Marshal(spec)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(out)
	return err
}
