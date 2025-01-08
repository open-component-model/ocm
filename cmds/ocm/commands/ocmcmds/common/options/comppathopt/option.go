package comppathopt

import (
	"github.com/spf13/pflag"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/common/output"
)

func From(o *output.Options) *Option {
	var opt *Option
	o.Get(&opt)
	return opt
}

func New() *Option {
	return &Option{}
}

type Option struct {
	Active bool
	Ids    []metav1.Identity
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Active, "path", "p", false, "follow component references")
}

// Complete consumes path identities if option is activated.
func (o *Option) Complete(args []string) ([]string, error) {
	var err error
	rest := args
	if o.Active {
		o.Ids, rest, err = common.ConsumeIdentities(false, args, ";")
	}
	return rest, err
}

func (o *Option) Usage() string {
	s := `
The <code>--path</code> options accepts a sequence of identities,
that will be used to follow component references a the specified
component(s).

In identity is given by a sequence of arguments starting with a
plain name value argument followed by any number of attribute assignments
of the form <code>&lt;<name>=&lt;value></code>.
The identity sequence stops at the end of the command line or with a sole
<code>;</code> argument, if other arguments are required for further purposes.
`
	return s
}
