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

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/options/closureoption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/processing"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Names = names.Components
	Verb  = commands.Get
)

type Command struct {
	utils.BaseCommand

	Output     output.Options
	Repository repooption.Option

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx), Output: *output.OutputOption(&closureoption.Option{
		AdditionalFields: output.Fields("IDENTITY"),
		FieldEnricher:    identity,
	})}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "get component version",
		Long: `
Get lists all component versions specified, if only a component is specified
all versions are listed.
` + o.Repository.Usage() + `
*Example:*
<pre>
$ ocm get componentversion ghcr.io/mandelsoft/kubelink
$ ocm get componentversion --repo OCIRegistry:ghcr.io mandelsoft/kubelink
</pre>
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.Repository.AddFlags(fs)
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	o.Refs = args
	if len(args) == 0 && o.Repository.Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	err := o.Repository.Complete(o.Context)
	if err != nil {
		return err
	}
	return o.Output.Complete(o.Context)

}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()
	repo, err := o.Repository.GetRepository(o.Context.OCM(), session)
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repo)
	return utils.HandleArgs(outputs, &o.Output, handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func identity(e interface{}) []string {
	p := e.(*comphdlr.Object)
	return []string{p.Identity.String()}
}

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) output.Output {
	def := &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", "PROVIDER", wide),
		Options: opts,
		Chain:   comphdlr.Sort,
		Mapping: mapping,
	}
	return closureoption.TableOutput(def, comphdlr.ClosureExplode).New()
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return TableOutput(opts, map_get_regular_output)
}

func get_wide(opts *output.Options) output.Output {
	return TableOutput(opts, map_get_wide_output, "REPOSITORY")
}

func map_get_regular_output(e interface{}) interface{} {
	p := e.(*comphdlr.Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	return []string{p.Spec.Component, tag, string(p.ComponentVersion.GetDescriptor().Provider)}
}

func map_get_wide_output(e interface{}) interface{} {
	p := e.(*comphdlr.Object)
	return output.Fields(map_get_regular_output(e), p.Spec.UniformRepositorySpec.String())
}
