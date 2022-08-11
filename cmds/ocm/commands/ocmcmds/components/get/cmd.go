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

package get

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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
		"component reference", output.Fields("IDENTITY"), options.Not(output.Selected("tree")), addIdentityField), schemaoption.New(""),
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

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
	"tree": get_tree,
}).AddChainedManifestOutputs(output.ComposeChain(closureoption.OutputChainFunction(comphdlr.ClosureExplode, comphdlr.Sort), Format))

func get_regular(opts *output.Options) output.Output {
	return TableOutput(opts, map_get_regular_output).New()
}

func get_wide(opts *output.Options) output.Output {
	return TableOutput(opts, map_get_wide_output, "REPOSITORY").New()
}

func get_tree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, map_get_regular_output), "NESTING").New()
}

func map_get_regular_output(e interface{}) interface{} {
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

func map_get_wide_output(e interface{}) interface{} {
	p := e.(*comphdlr.Object)
	return output.Fields(map_get_regular_output(e), p.Spec.UniformRepositorySpec.String())
}
