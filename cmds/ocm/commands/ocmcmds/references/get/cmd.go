// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	gcommon "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var (
	Names = names.References
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs, closureoption.New("component reference"), lookupoption.New()))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name> { <key>=<value> }}",
		Args:  cobra.MinimumNArgs(1),
		Short: "get references of a component version",
		Long: `
Get references of a component version. References are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.
`,
	}
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Comp = args[0]
	o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	return err
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	opts := output.From(o)
	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp})
	if err != nil {
		return err
	}
	return utils.HandleOutputs(opts, hdlr, utils.ElemSpecs(o.Ids)...)
}

////////////////////////////////////////////////////////////////////////////////

func reorder(it data.Iterable) data.Iterable {
	slice := elemhdlr.ObjectSlice(it)

outer:
	for i := 0; i < len(slice); i++ {
		o := slice[i]
		e := common.Elem(o)
		key := gcommon.NewNameVersion(e.ComponentName, e.Version)
		hist := o.GetHistory()
		nested := hist.Append(key)
		var j int
		for j = i + 1; j < len(slice); j++ {
			n := slice[j]
			if !n.GetHistory().HasPrefix(hist) {
				continue outer
			}
			if n.GetHistory().Equals(nested) {
				break
			}
		}
		o.Node = &key
		if j < len(slice) && j > i+1 {
			copy(slice[i:j-1], slice[i+1:j])
			slice[j-1] = o
		}
	}
	return slice
}

////////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields("NAME", "COMPONENT", "VERSION", wide),
		Options: opts,
		Chain:   elemhdlr.Sort.Transform(reorder),
		Mapping: mapping,
	}
}

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
	"tree": getTree,
}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetRegularOutput)).New()
}

func getWide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetWideOutput, "IDENTITY")).New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetWideOutput, "IDENTITY"), "COMPONENTVERSION").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	r := common.Elem(e)
	return output.Fields(r.GetName(), r.ComponentName, r.GetVersion())
}

func mapGetWideOutput(e interface{}) interface{} {
	o := e.(*elemhdlr.Object)
	return output.Fields(mapGetRegularOutput(e), o.Id.String())
}
