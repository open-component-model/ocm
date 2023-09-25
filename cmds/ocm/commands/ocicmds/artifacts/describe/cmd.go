// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package describe

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/handlers/artifacthdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	common2 "github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/ociutils"
	"github.com/open-component-model/ocm/pkg/generics"
	"github.com/open-component-model/ocm/pkg/out"
)

var (
	Names = names.Artifacts
	Verb  = verbs.Describe
)

func From(o *output.Options) *Options {
	var opt *Options
	o.Get(&opt)
	return opt
}

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
	utils.BaseCommand

	BlobFiles bool
	Refs      []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs, &Options{}))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artifact-reference>}",
		Short: "describe artifact version",
		Long: `
Describe lists all artifact versions specified, if only a repository is specified
all tagged artifacts are listed.
Per version a detailed, potentially recursive description is printed.

`,
		Example: `
$ ocm describe artifact ghcr.io/mandelsoft/kubelink
$ ocm describe artifact --repo OCIRegistry::ghcr.io mandelsoft/kubelink
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
	handler := artifacthdlr.NewTypeHandler(o.Context.OCI(), session, repooption.From(o).Repository)
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{}).AddChainedManifestOutputs(infoChain)

func getRegular(opts *output.Options) output.Output {
	return output.NewProcessingFunctionOutput(opts, processing.Chain(opts.LogContext()),
		generics.Conditional(From(opts).BlobFiles, outInfoWithFiles, outInfo))
}

func infoChain(options *output.Options) processing.ProcessChain {
	return processing.Chain(options.LogContext()).Parallel(4).Map(mapInfo(From(options)))
}

func outInfo(ctx out.Context, e interface{}) {
	p := e.(*artifacthdlr.Object)

	ociutils.PrintArtifact(common2.NewPrinter(ctx.StdOut()), p.Artifact, false)
}

func outInfoWithFiles(ctx out.Context, e interface{}) {
	p := e.(*artifacthdlr.Object)

	ociutils.PrintArtifact(common2.NewPrinter(ctx.StdOut()), p.Artifact, true)
}

type Info struct {
	Artifact string      `json:"ref"`
	Info     interface{} `json:"info"`
}

func mapInfo(opts *Options) processing.MappingFunction {
	return func(e interface{}) interface{} {
		p := e.(*artifacthdlr.Object)
		return &Info{
			Artifact: p.Spec.String(),
			Info:     ociutils.GetArtifactInfo(p.Artifact, opts.BlobFiles),
		}
	}
}
