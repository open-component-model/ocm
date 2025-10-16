package keyoption

import (
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm"
	ocmsign "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/signutils"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

func New() *Option {
	return &Option{}
}

type EvaluatedOptions struct {
	RootCerts signutils.GenericCertificatePool
	Keys      signing.KeyRegistry
}

type Option struct {
	ConfigFragment
	*EvaluatedOptions
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.ConfigFragment.AddFlags(fs)
}

func (o *Option) Configure(ctx ocm.Context) error {
	var err error
	o.EvaluatedOptions, err = o.ConfigFragment.Evaluate(ctx, nil)
	return err
}

func Usage() string {
	s := `
The <code>--public-key</code> and <code>--private-key</code> options can be
used to define public and private keys on the command line. The options have an
argument of the form <code>&lt;name>=&lt;filepath></code>. The name is the name
of the key and represents the context is used for (For example the signature
name of a component version)

Alternatively a key can be specified as base64 encoded string if the argument
start with the prefix <code>!</code> or as direct string with the prefix
<code>=</code>.

With <code>--issuer</code> it is possible to declare expected issuer 
constraints for public key certificates provided as part of a signature
required to accept the provisioned public key (besides the successful
validation of the certificate). By default, the issuer constraint is
derived from the signature name. If it is not a formal distinguished name,
it is assumed to be a plain common name.

With <code>--ca-cert</code> it is possible to define additional root
certificates for signature verification, if public keys are provided
by a certificate delivered with the signature.
`
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	opts.Keys = o.Keys
	opts.RootCerts = o.RootCerts
}
