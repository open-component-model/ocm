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
	"github.com/gardener/ocm/cmds/ocm/commands"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/names"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	ocmcommon "github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	compcommon "github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/components/common"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/components/common/options/repooption"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/resources/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
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
	Context clictx.Context
	Closure bool

	Repository repooption.Option
	Output     output.Options

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx}, names...)
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
	fs.BoolVarP(&o.Closure, "closure", "c", false, "follow component references")
	o.Repository.AddFlags(fs)
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	err := o.Output.Complete()
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
	repo, err := o.Repository.GetRepository(o.Context.OCM(), session)
	if err != nil {
		return err
	}
	vershdlr := compcommon.NewTypeHandler(o.Context.OCM(), session, repo)
	out := &output.SingleElementOutput{}
	err = utils.HandleOutput(out, vershdlr, utils.StringSpec(o.Comp))
	if err != nil {
		return err
	}
	hdlr := common.NewTypeHandler(repo, session, out.Elem.(*compcommon.Object).ComponentVersion, o.Closure)

	return utils.HandleOutputs(outputs, &o.Output, hdlr, utils.ElemSpecs(o.Ids)...)
}

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return output.NewProcessingTableOutput(opts, data.Chain().Map(map_get_regular_output),
		append(ocmcommon.MetaOutput, "TYPE", "RELATION")...)
}

func get_wide(opts *output.Options) output.Output {
	return output.NewProcessingTableOutput(opts, data.Chain().Map(map_get_wide_output),
		append(append(ocmcommon.MetaOutput, "TYPE", "RELATION"), ocmcommon.AccessOutput...)...)
}

func map_get_regular_output(e interface{}) interface{} {
	r := common.Elem(e)
	return append(ocmcommon.MapMetaOutput(e), r.Type, string(r.Relation))
}

func map_get_wide_output(e interface{}) interface{} {
	return append(map_get_regular_output(e).([]string), ocmcommon.MapAccessOutput(common.Elem(e).Access)...)
}
