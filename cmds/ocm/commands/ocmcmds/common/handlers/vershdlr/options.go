package vershdlr

import (
	"github.com/Masterminds/semver/v3"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/common/options"
)

type Option interface {
	ApplyToHandler(handler *TypeHandler)
}

type Options []Option

func (o Options) ApplyToHandler(handler *TypeHandler) {
	for _, e := range o {
		e.ApplyToHandler(handler)
	}
}

func OptionsFor(o options.OptionSetProvider) Options {
	var hopts []Option
	if constr := versionconstraintsoption.From(o); constr != nil {
		if len(constr.Constraints) > 0 {
			hopts = append(hopts, WithVersionConstraints(constr.Constraints))
		}
		if constr.Latest {
			hopts = append(hopts, LatestOnly())
		}
	}
	if lookup := lookupoption.From(o); lookup != nil {
		hopts = append(hopts, Resolver(lookup))
	}
	return hopts
}

////////////////////////////////////////////////////////////////////////////////

type constraints struct {
	constraints []*semver.Constraints
}

func (o constraints) ApplyToHandler(handler *TypeHandler) {
	handler.constraints = o.constraints
}

func WithVersionConstraints(c []*semver.Constraints) Option {
	return constraints{c}
}

////////////////////////////////////////////////////////////////////////////////

type latestonly struct {
	flag bool
}

func (o latestonly) ApplyToHandler(handler *TypeHandler) {
	handler.latest = o.flag
}

func LatestOnly(b ...bool) Option {
	return latestonly{utils.OptionalDefaultedBool(true, b...)}
}

////////////////////////////////////////////////////////////////////////////////

type resolver struct {
	resolver ocm.ComponentVersionResolver
}

func (o resolver) ApplyToHandler(handler *TypeHandler) {
	handler.resolver = o.resolver
}

func Resolver(r ocm.ComponentVersionResolver) Option {
	return resolver{r}
}

////////////////////////////////////////////////////////////////////////////////

type repository struct {
	repository ocm.Repository
}

func (o repository) ApplyToHandler(handler *TypeHandler) {
	handler.repobase = o.repository
}

func Repository(r ocm.Repository) Option {
	return repository{r}
}
