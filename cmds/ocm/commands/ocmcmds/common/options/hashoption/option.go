package hashoption

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/signingattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/normalizations/jsonv1"
	ocmsign "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
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
	fs.StringVarP(&o.NormAlgorithm, "normalization", "N", jsonv1.Algorithm, "normalization algorithm")
	fs.StringVarP(&o.hashAlgorithm, "hash", "H", sha256.Algorithm, "hash algorithm")
}

func (o *Option) Configure(ctx clictx.Context) error {
	if o.NormAlgorithm == "" {
		o.NormAlgorithm = jsonv1.Algorithm
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
` + listformat.FormatList(jsonv1.Algorithm, compdesc.Normalizations.Names()...)

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
