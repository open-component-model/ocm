package comppathopt

import (
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
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
The <code>--path</code> options accets a sequence of identities,
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
