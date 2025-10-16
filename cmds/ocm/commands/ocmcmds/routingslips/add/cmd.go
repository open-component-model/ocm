package add

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/spi"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

const (
	DEFAULT_CREDENTIALS_FILE = "TOICredentials"
	DEFAULT_PARAMETER_FILE   = "TOIParameters"
)

var (
	Names = names.RoutingSlips
	Verb  = verbs.Add
)

type Command struct {
	utils.BaseCommand
	CompSpec  string
	Name      string
	Type      string
	Links     []string
	Entry     *routingslip.GenericEntry
	Algorithm string
	Digest    string

	prov       flagsets.ExplicitlyTypedConfigTypeOptionSetConfigProvider
	configopts flagsets.ConfigOptions
}

// NewCommand creates a new routing slip add command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), lookupoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] <component-version> <routing-slip> <type>",
		Args:  cobra.ExactArgs(3),
		Short: "add routing slip entry",
		Long: `
Add a routing slip entry for the specified routing slip name to the given
component version. The name is typically a DNS domain name followed by some
qualifiers separated by a slash (/). It is possible to use arbitrary types,
the type is not checked, if it is not known. Accordingly, an arbitrary config
given as JSON or YAML can be given to determine the attribute set of the new
entry for unknown types.

` + routingslip.EntryUsage(spi.DefaultEntryTypeScheme(), true),
		Example: `
$ ocm add routingslip ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev mandelsoft.org comment --entry "comment=some text"
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
	// cmd.AddCommand(topicroutingslips.New(o.Context, "ocm-routingslips"))
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.prov = routingslip.For(o.OCMContext()).CreateConfigTypeSetConfigProvider()
	o.configopts = o.prov.CreateOptions()
	o.configopts.AddFlags(fs)

	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.Algorithm, "algorithm", "S", rsa.Algorithm, "signature handler")
	fs.StringVarP(&o.Digest, "digest", "", "", "parent digest to use")
	fs.StringSliceVarP(&o.Links, "links", "", nil, "links to other slip/entries (<slipname>[@<digest>])")
}

func (o *Command) Complete(args []string) error {
	o.CompSpec = args[0]
	o.Name = args[1]
	o.Type = args[2]

	if o.Type == "" {
		return errors.ErrInvalid(routingslip.KIND_ENTRY_TYPE, o.Type)
	}
	o.prov.SetTypeName(o.Type)

	data, err := o.prov.GetConfigFor(o.configopts)
	if err != nil {
		return err
	}
	u, err := runtime.ToUnstructuredTypedObject(data)
	if err != nil {
		return errors.Wrapf(err, "invalid entry data")
	}

	o.Entry = routingslip.AsGenericEntry(u)
	err = o.Entry.Validate(o.OCMContext())
	if err != nil {
		return err
	}
	if o.Algorithm == "" {
		o.Algorithm = rsa.Algorithm
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository)
	return utils.HandleOutput(&action{cmd: o}, handler, utils.StringElemSpecs(o.CompSpec)...)
}

////////////////////////////////////////////////////////////////////////////////

type action struct {
	data comphdlr.Objects
	cmd  *Command
}

var _ output.Output = (*action)(nil)

func (a *action) Add(e interface{}) error {
	if len(a.data) > 0 {
		return errors.New("found multiple component versions")
	}
	o, ok := e.(*comphdlr.Object)
	if !ok {
		return fmt.Errorf("object of type %T is not a valid comphdlr.Object", e)
	}
	a.data = append(a.data, o)
	return nil
}

func (a *action) Close() error {
	return nil
}

func (a *action) Out() error {
	if len(a.data) == 0 {
		return fmt.Errorf("no component version selected")
	}

	cv := a.data[0].ComponentVersion
	v, err := routingslip.Get(cv)
	if err != nil {
		return err
	}
	if v == nil {
		v = routingslip.LabelValue{}
	}

	var links []routingslip.Link
	for i, l := range a.cmd.Links {
		idx := strings.Index(l, "@")
		if idx <= 0 {
			if l == "all" {
				links = v.Leaves()
				break
			} else {
				slip, err := v.Query(l)
				if err != nil {
					return errors.ErrInvalid(routingslip.KIND_ROUTING_SLIP, l)
				}
				if slip != nil {
					for _, d := range slip.Leaves() {
						links = append(links, routingslip.Link{
							Name:   l,
							Digest: d,
						})
					}
				} else {
					return fmt.Errorf("link %q: slip not found", l)
				}
				continue
			}
		}
		n := l[:i]
		d := l[i+1:]
		slip, err := v.Query(n)
		if err != nil {
			return errors.ErrInvalid(routingslip.KIND_ROUTING_SLIP, n)
		}
		if slip == nil {
			return fmt.Errorf("link %q: slip %q not found", l, n)
		}
		var found digest.Digest
		for e := 0; e < slip.Len(); e++ {
			if strings.HasPrefix(slip.Get(e).Digest.Encoded(), d) {
				if found != "" {
					return fmt.Errorf("link %q: entry %q is not unique", l, d)
				}
				found = slip.Get(i).Digest
			}
		}
		if found == "" {
			return fmt.Errorf("link %q not found", l)
		}
		links = append(links, routingslip.Link{
			Name:   n,
			Digest: found,
		})
	}
	if a.cmd.Digest == "" {
		_, err = routingslip.AddEntry(cv, a.cmd.Name, a.cmd.Algorithm, a.cmd.Entry, links)
	} else {
		_, err = routingslip.AddEntry(cv, a.cmd.Name, a.cmd.Algorithm, a.cmd.Entry, links, digest.Digest(a.cmd.Digest))
	}
	if err != nil {
		return err
	}
	return cv.Update()
}
