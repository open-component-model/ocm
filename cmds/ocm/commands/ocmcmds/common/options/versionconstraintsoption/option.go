package versionconstraintsoption

import (
	"github.com/Masterminds/semver/v3"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/cobrautils/flag"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(silent ...bool) *Option {
	return &Option{SilentLatestOption: utils.Optional(silent...)}
}

type Option struct {
	SilentLatestOption bool
	Latest             bool
	Constraints        []*semver.Constraints
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	if !o.SilentLatestOption {
		fs.BoolVarP(&o.Latest, "latest", "", false, "restrict component versions to latest")
	}
	flag.SemverConstraintsVarP(fs, &o.Constraints, "constraints", "c", nil, "version constraint")
}

func (o *Option) SetLatest(latest ...bool) *Option {
	o.Latest = utils.OptionalDefaultedBool(true, latest...)
	return o
}

func (o *Option) Usage() string {
	s := `
If the option <code>--constraints</code> is given, and no version is specified
for a component, only versions matching the given version constraints
(semver https://github.com/Masterminds/semver) are selected.
`
	if !o.SilentLatestOption {
		s += `With <code>--latest</code> only
the latest matching versions will be selected.
`
	}
	return s
}
