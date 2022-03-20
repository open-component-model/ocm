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
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/components/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm"
	. "github.com/gardener/ocm/pkg/regex"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var TypeRegexp = Anchored(
	Optional(Sequence(Capture(Identifier), Literal("::"))),
	Capture(Match(".*")),
)

type Command struct {
	Context clictx.Context

	Output output.Options

	Repository string
	Refs       []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "get component version",
		Long: `
Get lists all component versions specified, if only a component is specified
all versions are listed.

If a <code>repo</code> option is specified the given names are interpreted 
as component names. 

The options follows the syntax [<repotype>::]<repospec>. The following
repository types are supported yet:
- <code>OCIRegistry</code>: The given repository spec is used as base url

Without a specified type prefix any JSON representation of an OCM repository
specification supported by the OCM library or the name of an OCM repository
configured in the used config file can be used.

*Example:*
<pre>
$ ocm get componentversion ghcr.io/mandelsoft/kubelink
$ ocm get componentversion --repo OCIRegistry:ghcr.io mandelsoft/kubelink
</pre>
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Repository, "repo", "r", "", "repository name or spec")
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	if len(args) == 0 && o.Repository == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Refs = args
	return nil
}

func (o *Command) Run() error {
	var err error
	var repobase ocm.Repository
	session := ocm.NewSession(nil)

	if o.Repository != "" {

		m := TypeRegexp.FindStringSubmatch(o.Repository)
		if m == nil {
			return errors.ErrInvalid("repository spec", o.Repository)
		}
		repobase, err = o.Context.OCM().DetermineRepository(m[1], m[2])
		if err != nil {
			return err
		}
	}

	handler := common.NewTypeHandler(o.Context.OCMContext(), session, repobase)

	return utils.HandleArgs(outputs, &o.Output, handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(get_regular, output.Outputs{
	// "wide": get_wide,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return output.NewProcessingTableOutput(opts, data.Chain().Map(map_get_regular_output),
		"REPOSITORY", "COMPONENT", "VERSION", "PROVIDER")
}

func map_get_regular_output(e interface{}) interface{} {
	p := e.(*common.Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	return []string{p.Spec.UniformRepositorySpec.String(), p.Spec.Component, tag, string(p.ComponentVersion.GetDescriptor().Provider)}
}
