package storeoption

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	ocmsign "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

var _ options.Options = (*Option)(nil)

func New(name ...string) *Option {
	return &Option{name: general.OptionalDefaulted("remember-verified", name...)}
}

type Option struct {
	name string

	// File is used to remember verify component versions.
	File                 string
	RememberVerification bool
	Store                ocmsign.VerifiedStore
}

const DEFAULT_VERIFIED_FILE = "~/.ocm/verified"

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.RememberVerification, o.name, false, "enable verification store")
	fs.StringVarP(&o.File, "verified", "", DEFAULT_VERIFIED_FILE, "file used to remember verifications for downloads")
}

func (o *Option) Configure(ctx clictx.Context) error {
	var err error
	if o.Store == nil && (o.RememberVerification || o.File != DEFAULT_VERIFIED_FILE) {
		o.Store, err = ocmsign.NewVerifiedStore(o.File, vfsattr.Get(ctx))
		if err != nil {
			return err
		}
	}
	if o.Store != nil {
		o.RememberVerification = true
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
If the verification store is enabled, resources downloaded from
signed or verified component versions are verified against their digests
provided by the component version.(not supported for using downloaders for the
resource download).

The usage of the verification store is enabled by <code>--` + o.name + `</code> or by
specifying a verification file with <code>--verified</code>.
`
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	opts.VerifiedStore = o.Store
}
