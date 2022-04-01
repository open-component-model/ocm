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
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands"
	ocmcommon "github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/options/closureoption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/processing"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/ocm"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	Names = names.Resources
	Verb  = commands.Get
)

type Command struct {
	utils.BaseCommand

	Repository repooption.Option
	Output     output.Options

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx), Output: *output.OutputOption(&closureoption.Option{}, &lookupoption.Option{})}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name> { <key>=<value> }}",
		Args:  cobra.MinimumNArgs(1),
		Short: "get resources of a component version",
		Long: `
Get resources of a component version. Sources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.Repository.AddFlags(fs)
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	err := o.Output.Complete(o.Context)
	if err != nil {
		return err
	}
	err = o.Repository.Complete(o.Context)
	if err != nil {
		return err
	}
	o.Comp = args[0]
	o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	return err
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.Output.ProcessOnOptions(ocmcommon.CompleteOptionsWithContext(o.OCM(), session))
	if err != nil {
		return err
	}
	err = o.Repository.CompleteWithSession(o.OCM(), session)
	if err != nil {
		return err
	}

	hdlr, err := common.NewTypeHandler(o.Context.OCM(), &o.Output, o.Repository.Repository, session, []string{o.Comp})
	if err != nil {
		return err
	}
	return utils.HandleOutputs(outputs, &o.Output, hdlr, utils.ElemSpecs(o.Ids)...)
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

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
	"tree": get_tree,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_regular_output)).New()
}

func get_wide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_wide_output, elemhdlr.AccessOutput...)).New()
}

func get_tree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, map_get_regular_output), "NESTING").New()
}

func map_get_regular_output(e interface{}) interface{} {
	r := common.Elem(e)
	return append(elemhdlr.MapMetaOutput(e), r.Type, string(r.Relation))
}

func map_get_wide_output(e interface{}) interface{} {
	return output.Fields(map_get_regular_output(e), elemhdlr.MapAccessOutput(common.Elem(e).Access))
}
