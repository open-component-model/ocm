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
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var (
	Names = names.Resources
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, versionconstraintsoption.New(), repooption.New(), output.OutputOptions(outputs, closureoption.New("component reference"), lookupoption.New()))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name> { <key>=<value> }}",
		Args:  cobra.MinimumNArgs(1),
		Short: "get resources of a component version",
		Long: `
Get resources of a component version. Resources are specified
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
	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp}, elemhdlr.ForceEmpty(output.Selected("tree")(opts)), common.OptionsFor(o))
	if err != nil {
		return err
	}
	return utils.HandleOutputs(opts, hdlr, utils.ElemSpecs(o.Ids)...)
}

////////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields(elemhdlr.MetaOutput, "TYPE", "RELATION", wide),
		Options: opts,
		Chain:   elemhdlr.Sort,
		Mapping: mapping,
	}
}

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide":     getWide,
	"tree":     getTree,
	"treewide": getTreewide,
}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetRegularOutput)).New()
}

func getWide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetWideOutput, elemhdlr.AccessOutput...)).New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetRegularOutput), "COMPONENT", elemhdlr.NodeMapping(2), output.TreeNodeTitleFunc(elemhdlr.NodeTitle)).New()
}

func getTreewide(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetTreewideOutput, "ACCESS"), "COMPONENT", elemhdlr.NodeMapping(3), output.TreeNodeTitleFunc(elemhdlr.NodeTitle)).New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	r := common.Elem(e)
	return append(elemhdlr.MapMetaOutput(e), r.Type, string(r.Relation))
}

func mapGetWideOutput(e interface{}) interface{} {
	return output.Fields(mapGetRegularOutput(e), elemhdlr.MapAccessOutput(common.Elem(e).Access))
}

func mapGetTreewideOutput(e interface{}) interface{} {
	return output.Fields(mapGetRegularOutput(e), common.Elem(e).Access.GetKind())
}
