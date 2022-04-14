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

	"github.com/open-component-model/ocm/cmds/ocm/commands"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/handlers/artefacthdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/oci"
	"github.com/spf13/cobra"
)

var (
	Names = names.Artefacts
	Verb  = commands.Get
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new artefact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, &repooption.Option{}, output.OutputOptions(outputs, closureoption.New("index")))}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artefact-reference>}",
		Short: "get artefact version",
		Long: `
Get lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
	`,
		Example: `
$ ocm get artefact ghcr.io/mandelsoft/kubelink
$ ocm get artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
`,
	}
}

func (o *Command) Complete(args []string) error {
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Refs = args
	return nil
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()
	err := o.ProcessOnOptions(common.CompleteOptionsWithContext(o.Context, session))
	if err != nil {
		return err
	}
	handler := artefacthdlr.NewTypeHandler(o.Context.OCI(), session, repooption.From(o).Repository)
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields("REGISTRY", "REPOSITORY", "KIND", "TAG", "DIGEST", wide),
		Chain:   artefacthdlr.Sort,
		Options: opts,
		Mapping: mapping,
	}
}

var outputs = output.NewOutputs(get_regular, output.Outputs{
	"wide": get_wide,
}).AddManifestOutputs()

func get_regular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_regular_output), artefacthdlr.ClosureExplode).New()
}

func get_wide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, map_get_wide_output, "MIMETYPE", "CONFIGTYPE"), artefacthdlr.ClosureExplode).New()
}

func map_get_regular_output(e interface{}) interface{} {
	digest := "unknown"
	p := e.(*artefacthdlr.Object)
	blob, err := p.Artefact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.Spec.Tag != nil {
		tag = *p.Spec.Tag
	}
	kind := "-"
	if p.Artefact.IsManifest() {
		kind = "manifest"
	}
	if p.Artefact.IsIndex() {
		kind = "index"
	}
	return []string{p.Spec.Host, p.Spec.Repository, kind, tag, digest}
}

func map_get_wide_output(e interface{}) interface{} {
	p := e.(*artefacthdlr.Object)

	config := "-"
	if p.Artefact.IsManifest() {
		config = p.Artefact.ManifestAccess().GetDescriptor().Config.MediaType
	}
	return output.Fields(map_get_regular_output(e), p.Artefact.GetDescriptor().MimeType(), config)
}
