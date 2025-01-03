package hashoption

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/jsonv1"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/jsonv2"
	"ocm.software/ocm/api/ocm/compdesc/normalizations/jsonv3"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	ocmsign "ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/utils/listformat"
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

type Option struct {
	Hasher        signing.Hasher
	NormAlgorithm string
	hashAlgorithm string
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.NormAlgorithm, "normalization", "N", jsonv3.Algorithm, "normalization algorithm")
	fs.StringVarP(&o.hashAlgorithm, "hash", "H", sha256.Algorithm, "hash algorithm")
}

func (o *Option) Configure(ctx clictx.Context) error {
	if o.NormAlgorithm == "" {
		o.NormAlgorithm = jsonv3.Algorithm
	}
	if o.hashAlgorithm == "" {
		o.hashAlgorithm = sha256.Algorithm
	}
	x := compdesc.Normalizations.Get(o.NormAlgorithm)
	if x == nil {
		return errors.ErrUnknown(compdesc.KIND_NORM_ALGORITHM, o.NormAlgorithm)
	}
	o.Hasher = signingattr.Get(ctx).GetHasher(o.hashAlgorithm)
	if o.Hasher == nil {
		return errors.ErrUnknown(compdesc.KIND_HASH_ALGORITHM, o.hashAlgorithm)
	}
	return nil
}

func (o *Option) Usage() string {
	s := `
The following normalization modes are supported with option <code>--normalization</code>:
` + listformat.FormatList(jsonv3.Algorithm, compdesc.Normalizations.Names()...)

	s += `

Note that the normalization algorithm is important to be equivalent when used for signing and verification, otherwise
the verification can fail. Please always migrate to the latest normalization algorithm whenever possible.
New signature algorithms can be used as soon as they are available in the component version after signing it.

The algorithms ` + jsonv1.Algorithm + ` and ` + jsonv2.Algorithm + ` are deprecated and should not be used anymore.
Please switch to ` + jsonv3.Algorithm + ` as soon as possible.

`

	s += `

The following hash modes are supported with option <code>--hash</code>:
` + listformat.FormatList(sha256.Algorithm, signing.DefaultRegistry().HasherNames()...)

	signing.DefaultRegistry().HasherNames()
	return s
}

var _ ocmsign.Option = (*Option)(nil)

func (o *Option) ApplySigningOption(opts *ocmsign.Options) {
	opts.NormalizationAlgo = o.NormAlgorithm
	opts.Hasher = o.Hasher
}
