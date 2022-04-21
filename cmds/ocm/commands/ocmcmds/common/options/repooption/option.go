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

package repooption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/runtime"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	Spec       string
	Repository ocm.Repository
}

var _ common.OptionCompleter = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Spec, "repo", "r", "", "repository name or spec")
}

func (o *Option) CompleteWithSession(octx clictx.OCM, session ocm.Session) error {
	if o.Repository == nil {
		r, err := o.GetRepository(octx, session)
		if err != nil {
			return err
		}
		o.Repository = r
	}
	return nil
}

func (o *Option) GetRepository(ctx clictx.OCM, session ocm.Session) (ocm.Repository, error) {
	if o.Spec != "" {
		r, _, err := session.DetermineRepository(ctx.Context(), o.Spec, ctx.GetAlias)
		return r, err
	}
	return nil, nil
}

func (o *Option) Usage() string {
	s := `
If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center><code>&lt;component>[:&lt;version>]</code></center>

If no <code>--repo</code> option is specified the given names are interpreted 
as located OCM component version references:

<center><code>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</code></center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center><code>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</code></center>

The <code>--repo</code> option takes an OCM repository specification:

<center><code>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></code></center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository type supported by the 
linked library can be used:

Dedicated OCM repository types:
`

	types := runtime.KindNames(ocm.DefaultContext().RepositoryTypes())
	for _, t := range types {
		s += "- `" + t + "`\n"
	}

	s += `
OCI Repository types (using standard component repository to OCI mapping):
`
	types = runtime.KindNames(oci.DefaultContext().RepositoryTypes())
	for _, t := range types {
		s += "- `" + t + "`\n"
	}
	return s
}
