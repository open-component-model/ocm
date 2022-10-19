// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

var (
	Names = names.Components
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs, closureoption.New(
		"component reference", output.Fields("IDENTITY"), options.Not(output.Selected("tree")), addIdentityField), lookupoption.New(), schemaoption.New(""),
	))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "get component version",
		Long: `
Get lists all component versions specified, if only a component is specified
all versions are listed.
`,
		Example: `
$ ocm get componentversion ghcr.io/mandelsoft/kubelink
$ ocm get componentversion --repo OCIRegistry:ghcr.io mandelsoft/kubelink
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.Refs = args
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
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
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func addIdentityField(e interface{}) []string {
	p := e.(*comphdlr.Object)
	return []string{p.Identity.String()}
}

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", "PROVIDER", wide),
		Options: opts,
		Chain:   comphdlr.Sort,
		Mapping: mapping,
	}
	return closureoption.TableOutput(def, comphdlr.ClosureExplode)
}

/////////////////////////////////////////////////////////////////////////////

func Format(opts *output.Options) processing.ProcessChain {
	o := schemaoption.From(opts)
	if o.Schema == "" {
		return nil
	}
	return processing.Map(func(in interface{}) interface{} {
		desc := comphdlr.Elem(in).GetDescriptor()
		out, err := compdesc.Convert(desc, compdesc.SchemaVersion(o.Schema))
		if err != nil {
			return struct {
				Scheme  string `json:"scheme"`
				Name    string `json:"name"`
				Version string `json:"version"`
				Error   string `json:"error"`
			}{
				Scheme:  desc.SchemaVersion(),
				Name:    desc.GetName(),
				Version: desc.GetVersion(),
				Error:   err.Error(),
			}
		} else {
			return out
		}
	})
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
	"tree": getTree,
}).AddChainedManifestOutputs(output.ComposeChain(closureoption.OutputChainFunction(comphdlr.ClosureExplode, comphdlr.Sort), Format))

func getRegular(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetRegularOutput).New()
}

func getWide(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetWideOutput, "REPOSITORY").New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetRegularOutput), "NESTING").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := e.(*comphdlr.Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	if p.ComponentVersion == nil {
		return []string{p.Spec.Component, tag, "<unknown component version>"}
	}
	return []string{p.Spec.Component, tag, string(p.ComponentVersion.GetDescriptor().Provider.Name)}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := e.(*comphdlr.Object)
	return output.Fields(mapGetRegularOutput(e), p.Spec.UniformRepositorySpec.String())
}
