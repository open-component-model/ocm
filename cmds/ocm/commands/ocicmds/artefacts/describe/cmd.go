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

package describe

import (
	"fmt"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands/ocicmds/artefacts/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/data"
	"github.com/gardener/ocm/cmds/ocm/pkg/output"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/oci"
	"github.com/gardener/ocm/pkg/oci/ociutils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Options struct {
	BlobFiles bool
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.BlobFiles, "layerfiles", "", false, "list layer files")
}

func (o *Options) Complete() error {
	return nil
}

type Command struct {
	Context clictx.Context

	Output output.Options

	BlobFiles  bool
	Repository common.RepositoryOptions
	Refs       []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx, Output: output.Options{Others: &Options{}}}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artefact-reference>}",
		Short: "describe artefact version",
		Long: `
Describe lists all artefact versions specified, if only a repository is specified
all tagged artefacts are listed.
Per version a detailed, potentially recursive description is printed.

` + o.Repository.Usage() + `

*Example:*
<pre>
$ ocm describe artefact ghcr.io/mandelsoft/kubelink
$ ocm describe artefact --repo OCIRegistry:ghcr.io mandelsoft/kubelink
</pre>
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.Repository.AddFlags(fs)
	o.Output.AddFlags(fs, outputs)
}

func (o *Command) Complete(args []string) error {
	var err error
	if len(args) == 0 && o.Repository.Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Refs = args
	err = o.Repository.Complete(o.Context)
	if err != nil {
		return err
	}

	return nil
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()
	handler := common.NewTypeHandler(o.Context.OCIContext(), session, o.Repository.Repository)
	session.Closer(handler, nil)
	return utils.HandleArgs(outputs, &o.Output, handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(get_regular, output.Outputs{}).AddChainedManifestOutputs(infoChain)

func get_regular(opts *output.Options) output.Output {
	return output.NewProcessingFunctionOutput(data.Chain(), outInfo)
}

func infoChain(options *output.Options) data.ProcessChain {
	return data.Chain().Parallel(4).Map(mapInfo(options.Others.(*Options)))
}

func outInfo(e interface{}) {
	p := e.(*common.Object)
	fmt.Printf("%s", ociutils.PrintArtefact(p.Artefact))
}

type Info struct {
	Artefact string      `json:"artefact"`
	Info     interface{} `json:"info"`
}

func mapInfo(opts *Options) data.MappingFunction {
	return func(e interface{}) interface{} {
		p := e.(*common.Object)
		return &Info{
			Artefact: p.Spec.String(),
			Info:     ociutils.GetArtefactInfo(p.Artefact, opts.BlobFiles),
		}
	}
}
