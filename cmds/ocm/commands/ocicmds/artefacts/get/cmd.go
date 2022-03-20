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
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/artefacts/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
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
		Use:   "[<options>] {<artefact-reference>}",
		Short: "get artefact version",
		Long: `
Get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.

If a <code>repo</code> option is specified the given names are interpreted 
as located OCI artefact names. 

The options follows the syntax [<repotype>::]<repospec>. The following
repository types are supported yet:
- <code>OCIRegistry</code>: The given repository spec is used as base url
- <code>Docker</code>/<code>DockerDeamon</code>: The repository spec is option and used as daemon address

Without a specified type prefix any JSON representation of an OCI repository
specification supported by the OCM library or the name of an OCI repository
configured in the used config file can be used.

If the repository option is specified, the given artefact names are interpreted
relative to the specified repository.

*Example:*
<pre>
$ ocm get artefact ghcr.io/mandelsoft/kubelink
$ ocm get artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
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
	var repobase oci.Repository
	session := oci.NewSession(nil)

	if o.Repository != "" {

		m := TypeRegexp.FindStringSubmatch(o.Repository)
		if m == nil {
			return errors.ErrInvalid("repository spec", o.Repository)
		}
		repobase, err = o.Context.OCI().DetermineRepository(m[1], m[2])
		if err != nil {
			return err
		}
	}

	handler := common.NewTypeHandler(o.Context.OCIContext(), session, repobase)

	return utils.HandleArgs(outputs, &o.Output, handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return output.NewProcessingTableOutput(opts, data.Chain().Map(map_get_regular_output),
		"REGISTRY", "REPOSITORY", "TAG", "DIGEST")
}

func get_wide(opts *output.Options) output.Output {
	return output.NewProcessingTableOutput(opts, data.Chain().Parallel(20).Map(map_get_wide_output),
		"REGISTRY", "REPOSITORY", "TAG", "DIGEST", "MIMETYPE", "CONFIGTYPE")
}

func map_get_regular_output(e interface{}) interface{} {
	digest := "unknown"
	p := e.(*common.Object)
	blob, err := p.Artefact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.Spec.Tag != nil {
		tag = *p.Spec.Tag
	}
	return []string{p.Spec.Host, p.Spec.Repository, tag, digest}
}

func map_get_wide_output(e interface{}) interface{} {
	digest := "unknown"
	p := e.(*common.Object)
	blob, err := p.Artefact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.Spec.Tag != nil {
		tag = *p.Spec.Tag
	}
	config := "-"
	if p.Artefact.IsManifest() {
		config = p.Artefact.ManifestAccess().GetDescriptor().Config.MediaType
	}
	return []string{p.Spec.Host, p.Spec.Repository, tag, digest, p.Artefact.GetDescriptor().MimeType(), config}
}
