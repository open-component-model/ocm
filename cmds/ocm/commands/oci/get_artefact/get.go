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

package get_artefact

import (
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/cmd"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

type Options struct {
	Context cmd.Context

	Output output.Options

	// Ref is the oci artifact reference.
	Repository string

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx cmd.Context) *cobra.Command {
	opts := &Options{Context: ctx}
	cmd := &cobra.Command{
		Use:              "artefact",
		TraverseChildren: true,
		Aliases:          []string{"a", "art"},
		Short:            "get artefact version",
		Long: `
get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.Run(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())
	return cmd
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Repository, "repo", "r", "", "repository name or spec")
	o.Output.AddFlags(fs)
}

func (o *Options) Complete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one argument that defines the reference is needed")
	}
	o.Refs = args
	return nil
}

func (o *Options) Run() error {
	var repobase oci.Repository
	session := oci.NewSession()

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

	handler := &TypeHandler{
		octx:     o.Context.OCIContext(),
		session:  session,
		repobase: repobase,
	}

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
	p := e.(*Object)
	blob, err := p.artefact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.spec.Tag != nil {
		tag = *p.spec.Tag
	}
	return []string{p.spec.Host, p.spec.Repository, tag, digest}
}

func map_get_wide_output(e interface{}) interface{} {
	digest := "unknown"
	p := e.(*Object)
	blob, err := p.artefact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.spec.Tag != nil {
		tag = *p.spec.Tag
	}
	config := "-"
	if p.artefact.IsManifest() {
		config = p.artefact.ManifestAccess().GetDescriptor().Config.MediaType
	}
	return []string{p.spec.Host, p.spec.Repository, tag, digest, p.artefact.GetDescriptor().MimeType(), config}
}
