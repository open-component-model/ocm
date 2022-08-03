// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package lookupoption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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
			resolvers = append(resolvers, ocm.NewSessionBasedResolver(session, r))
		}
		return ocm.NewCompoundResolver(resolvers...), nil
	}
	return nil, nil
}

func (o *Option) Usage() string {
	s := `
If a component lookup for building a reference closure is required
the <code>--lookup</code>  option can be used to specify a fallback
lookup repository. 
By default the component versions are searched in the repository
holding the component version for which the closure is determined.
For *Component Archives* this is never possible, because it only
contains a single component version. Therefore, in this scenario
this option must always be specified to be able to follow component
references.
`
	return s
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
