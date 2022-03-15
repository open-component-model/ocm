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

	"github.com/gardener/ocm/cmds/ocm/cmd"
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/artefact"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

type Command struct {
	Context cmd.Context

	Output output.Options

	Repository string
	Refs       []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx cmd.Context) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx},
		&cobra.Command{
			Use:     "artefact[<options>] {<artefact-reference>}",
			Aliases: []string{"a", "art"},
			Short:   "get artefact version",
			Long: `
Get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
`,
		})
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Repository, "repo", "r", "", "repository name or spec")
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one argument that defines the reference is needed")
	}
	o.Refs = args
	return nil
}

func (o *Command) Run() error {
	var repobase oci.Repository
	session := oci.NewSession(nil)

	if o.Repository != "" {
		var parsed interface{}
		err := yaml.Unmarshal([]byte(o.Repository), &parsed)
		if err != nil {
			return errors.Wrapf(err, "cannot unmarshal repository spec")
		}
		if s, ok := parsed.(string); ok {
			repobase, err = o.Context.GetOCIRepository(s)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("spec: %s\n", o.Repository)
		}
	}

	handler := artefact.NewTypeHandler(o.Context.OCIContext(), session, repobase)

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
	p := e.(*artefact.Object)
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
	p := e.(*artefact.Object)
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
