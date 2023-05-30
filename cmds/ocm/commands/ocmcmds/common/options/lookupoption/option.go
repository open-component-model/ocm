// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package lookupoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer/transferhandler/standard"
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
	RepoSpecs []string
	Resolver  ocm.ComponentVersionResolver
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringArrayVarP(&o.RepoSpecs, "lookup", "", nil, "repository name or spec for closure lookup fallback")
}

func (o *Option) CompleteWithSession(octx clictx.OCM, session ocm.Session) error {
	if len(o.RepoSpecs) != 0 && o.Resolver == nil {
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
		resolvers := []ocm.ComponentVersionResolver{}
		for _, s := range o.RepoSpecs {
			r, _, err := session.DetermineRepository(ctx.Context(), s)
			if err != nil {
				return nil, err
			}
			resolvers = append(resolvers, r)
		}
		return ocm.NewCompoundResolver(resolvers...), nil
	}
	return nil, nil
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

func (o *Option) ApplyTransferOption(opts transferhandler.TransferOptions) error {
	return standard.Resolver(o.Resolver).ApplyTransferOption(opts)
}
