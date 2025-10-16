package lookupoption

import (
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/resolvers"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/cmds/ocm/common/options"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	standard.TransferOptionsCreator
	RepoSpecs []string
	Resolver  ocm.ComponentVersionResolver
}

var (
	_ transferhandler.TransferOption = (*Option)(nil)
	_ ocm.ComponentVersionResolver   = (*Option)(nil)
	_ ocm.ComponentResolver          = (*Option)(nil)
)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.RepoSpecs, "lookup", "", nil, "repository name or spec for closure lookup fallback")
}

func (o *Option) CompleteWithSession(octx clictx.OCM, session ocm.Session) error {
	if o.Resolver == nil {
		r, err := o.getResolver(octx, session)
		if err != nil {
			return err
		}
		o.Resolver = r
	}
	return nil
}

func (o *Option) getResolver(ctx clictx.OCM, session ocm.Session) (ocm.ComponentVersionResolver, error) {
	if len(o.RepoSpecs) != 0 {
		resolver := []ocm.ComponentVersionResolver{}
		for _, s := range o.RepoSpecs {
			r, _, err := session.DetermineRepository(ctx.Context(), s)
			if err != nil {
				return nil, err
			}
			resolver = append(resolver, r)
		}
		return resolvers.NewCompoundResolver(append(resolver, ctx.Context().GetResolver())...), nil
	}
	return ctx.Context().GetResolver(), nil
}

func (o *Option) Usage() string {
	s := `\
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. By default, the component versions are searched in
the repository holding the component version for which the closure is
determined. For *Component Archives* this is never possible, because
it only contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.
`
	return s
}

func (o *Option) IsGiven() bool {
	return len(o.RepoSpecs) > 0
}

func (o *Option) LookupComponentVersion(name string, vers string) (ocm.ComponentVersionAccess, error) {
	if o == nil || o.Resolver == nil {
		return nil, nil
	}
	cv, err := o.Resolver.LookupComponentVersion(name, vers)
	if err != nil {
		return nil, err
	}
	return cv, err
}

func (o *Option) LookupComponentProviders(name string) []ocm.ResolvedComponentProvider {
	if o != nil && o.Resolver != nil {
		if c, ok := o.Resolver.(ocm.ComponentResolver); ok {
			return c.LookupComponentProviders(name)
		}
	}
	return nil
}

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	return standard.Resolver(o.Resolver).ApplyTransferOption(opts)
}
