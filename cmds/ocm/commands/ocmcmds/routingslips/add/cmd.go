// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/opencontainers/go-digest"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/cobrautils/flagsets"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/spi"
	"github.com/open-component-model/ocm/pkg/errors"
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
	}
	// cmd.AddCommand(topicroutingslips.New(o.Context, "ocm-routingslips"))
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.prov = routingslip.For(o.OCMContext()).CreateConfigTypeSetConfigProvider().(flagsets.ExplicitlyTypedConfigTypeOptionSetConfigProvider)
	o.configopts = o.prov.CreateOptions()
	o.configopts.AddFlags(fs)

	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.Algorithm, "algorithm", "S", rsa.Algorithm, "signature handler")
	fs.StringVarP(&o.Digest, "digest", "", "", "parent digest to use")
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

	var err error
	if a.cmd.Digest == "" {
		_, err = routingslip.AddEntry(a.data[0].ComponentVersion, a.cmd.Name, a.cmd.Algorithm, a.cmd.Entry)
	} else {
		_, err = routingslip.AddEntry(a.data[0].ComponentVersion, a.cmd.Name, a.cmd.Algorithm, a.cmd.Entry, digest.Digest(a.cmd.Digest))
	}
	return err
}
