package verify

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
	Name = "verify"
)

func New(p ppi.Plugin) *cobra.Command {
	opts := Options{}

	cmd := &cobra.Command{
		Use:   Name + " <name> <digest> <signature>",
		Short: "verify signature",
		Long: `
Verify the given digest and signature with signing context taken from *stdin*.`,
		Args: cobra.ExactArgs(3),
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
	Name      string
	Digest    string
	Signature *signing.Signature
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
}

func (o *Options) Complete(args []string) error {
	var sig ppi.SignatureSpec

	err := json.Unmarshal([]byte(args[2]), &sig)
	if err != nil {
		return err
	}

	o.Name = args[0]
	o.Digest = args[1]
	o.Signature = sig.ConvertToSigning()
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
	v := h.GetVerifier()
	if v == nil {
		return fmt.Errorf("handler %q does not support verifying", opts.Name)
	}

	sctx := &signing.DefaultSigningContext{
		Hash:       ctx.HashAlgo,
		PrivateKey: ctx.PrivateKey,
		PublicKey:  ctx.PublicKey,
		Issuer:     ctx.Issuer,
	}

	return v.Verify(opts.Digest, opts.Signature, sctx)
}
