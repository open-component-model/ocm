package repooption

import (
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/utils/listformat"
	"ocm.software/ocm/api/utils/runtime"
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
	Spec       string
	Repository oci.Repository
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Spec, "repo", "", "", "repository name or spec")
}

func (o *Option) Configure(ctx clictx.Context) error {
	return nil
}

func (o *Option) CompleteWithSession(octx clictx.OCI, session oci.Session) error {
	if o.Repository == nil {
		r, err := o.GetRepository(octx, session)
		if err != nil {
			return err
		}
		o.Repository = r
	}
	return nil
}

func (o *Option) GetRepository(ctx clictx.OCI, session oci.Session) (oci.Repository, error) {
	if o.Spec != "" {
		r, _, err := session.DetermineRepository(ctx.Context(), o.Spec)
		return r, err
	}
	return nil, nil
}

func (o *Option) Usage() string {
	s := `
If the repository/registry option is specified, the given names are interpreted
relative to the specified registry using the syntax

<center>
    <pre>&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

If no <code>--repo</code> option is specified the given names are interpreted 
as extended OCI artifact references.

<center>
    <pre>[&lt;repo type>::]&lt;host>[:&lt;port>]/&lt;OCI repository name>[:&lt;tag>][@&lt;digest>]</pre>
</center>

The <code>--repo</code> option takes a repository/OCI registry specification:

<center>
    <pre>[&lt;repo type>::]&lt;configured name>|&lt;file path>|&lt;spec json></pre>
</center>

For the *Common Transport Format* the types <code>directory</code>,
<code>tar</code> or <code>tgz</code> are possible.

Using the JSON variant any repository types supported by the 
linked library can be used:
`
	return s + listformat.FormatMapElements("", runtime.KindToVersionList(oci.DefaultContext().RepositoryTypes().KnownTypeNames()))
}
