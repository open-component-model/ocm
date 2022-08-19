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

	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references/common"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/data"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	gcommom "github.com/open-component-model/ocm/pkg/common"
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
	specs, err := utils.ElemSpecs(o.Ids)
	if err != nil {
		return fmt.Errorf("failed to get elem specs: %w", err)
	}
	return utils.HandleOutputs(opts, hdlr, specs...)
}

////////////////////////////////////////////////////////////////////////////////

func reorder(it data.Iterable) data.Iterable {
	slice := elemhdlr.ObjectSlice(it)

outer:
	for i := 0; i < len(slice); i++ {
		o := slice[i]
		e := common.Elem(o)
		key := gcommom.NewNameVersion(e.ComponentName, e.Version)
		hist := o.GetHistory()
		nested := append(hist, key)
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

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
	"tree": get_tree,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_regular_output)).New()
}

func get_wide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_wide_output, "IDENTITY")).New()
}

func get_tree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, map_get_wide_output, "IDENTITY"), "COMPONENTVERSION").New()
}

func map_get_regular_output(e interface{}) interface{} {
	r := common.Elem(e)
	return output.Fields(r.GetName(), r.ComponentName, r.GetVersion())
}

func map_get_wide_output(e interface{}) interface{} {
	o := e.(*elemhdlr.Object)
	return output.Fields(map_get_regular_output(e), o.Id.String())
}
