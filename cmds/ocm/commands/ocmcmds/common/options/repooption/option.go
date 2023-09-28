// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repooption

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/listformat"
	"github.com/open-component-model/ocm/pkg/runtime"
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
	Spec       string
	Repository ocm.Repository
}

var _ common.OptionWithSessionCompleter = (*Option)(nil)

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Spec, "repo", "", "", "repository name or spec")
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
		r, _, err := session.DetermineRepository(ctx.Context(), o.Spec)
		return r, err
	}
	return nil, nil
}

func (o *Option) Usage() string {
	s := `
If the <code>--repo</code> option is specified, the given names are interpreted
relative to the specified repository using the syntax

<center>
    <pre>&lt;component>[:&lt;version>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as located OCM component version references:

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>][/&lt;base path>]//&lt;component>[:&lt;version>]</pre>
</center>

Additionally there is a variant to denote common transport archives
and general repository specifications

<center>
    <pre>[&lt;repo type>::]&lt;filepath>|&lt;spec json>[//&lt;component>[:&lt;version>]]</pre>
</center>

The <code>--repo</code> option takes an OCM repository specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> is possible.

Using the JSON variant any repository types supported by the 
linked library can be used:

Dedicated OCM repository types:
`

	s += listformat.FormatMapElements("", runtime.KindToVersionList(ocm.DefaultContext().RepositoryTypes().KnownTypeNames()))

	s += `
OCI Repository types (using standard component repository to OCI mapping):
`
	s += listformat.FormatMapElements("", runtime.KindToVersionList(oci.DefaultContext().RepositoryTypes().KnownTypeNames(), genericocireg.Excludes...))
	return s
}
