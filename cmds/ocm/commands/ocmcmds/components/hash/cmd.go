// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/hashoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
)

var (
	Names = names.Components
	Verb  = verbs.Hash
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{BaseCommand: utils.NewBaseCommand(ctx,
			versionconstraintsoption.New(),
			output.OutputOptions(outputs, &Option{}, closureoption.New(
				"component reference", output.Fields("IDENTITY"), addIdentityField), lookupoption.New(), hashoption.New(), repooption.New(),
			))},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "hash component version",
		Long: `
Hash lists normalized forms for all component versions specified, if only a component is specified
all versions are listed.
`,
		Example: `
$ ocm hash componentversion ghcr.io/mandelsoft/kubelink
$ ocm hash componentversion --repo OCIRegistry:ghcr.io mandelsoft/kubelink
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.Refs = args
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	err = From(o).Complete(o)
	if err != nil {
		return err
	}

	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, comphdlr.OptionsFor(o))
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func addIdentityField(e interface{}) []string {
	p := e.(*comphdlr.Object)
	return []string{p.Identity.String()}
}

func TableOutput(opts *output.Options, h *handler, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", "HASH", wide),
		Options: opts,
		Chain:   comphdlr.Sort.Map(h.digester),
		Mapping: mapping,
	}
	return closureoption.TableOutput(def, comphdlr.ClosureExplode)
}

////////////////////////////////////////////////////////////////////////////////

type Object struct {
	Spec       ocm.RefSpec
	History    common.History
	Descriptor *compdesc.ComponentDescriptor
	Error      error
}

type Manifest struct {
	History    common.History `json:"context"`
	Component  string         `json:"component"`
	Version    string         `json:"version"`
	Normalized string         `json:"normalized,omitempty"`
	Hash       string         `json:"hash,omitempty"`
	Error      string         `json:"error,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
}).AddChainedManifestOutputs(output.ComposeChain(closureoption.OutputChainFunction(comphdlr.ClosureExplode, comphdlr.Sort), mapManifest))

func mapManifest(opts *output.Options) processing.ProcessChain {
	h := newHandler(opts)
	return processing.Map(h.digester).Map(h.manifester)
}

func getRegular(opts *output.Options) output.Output {
	h := newHandler(opts)
	return TableOutput(opts, h, h.mapGetRegularOutput).New()
}

func getWide(opts *output.Options) output.Output {
	h := newHandler(opts)
	return TableOutput(opts, h, h.mapGetWideOutput, "NORMALIZED FORM").New()
}

////////////////////////////////////////////////////////////////////////////////

type handler struct {
	opts *hashoption.Option
	mode *Option
}

func newHandler(opts *output.Options) *handler {
	return &handler{
		opts: hashoption.From(opts),
		mode: From(opts),
	}
}

func (h *handler) manifester(e interface{}) interface{} {
	p := e.(*Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}

	hist := p.History
	if hist == nil {
		hist = common.History{}
	}

	m := &Manifest{
		History:   hist,
		Version:   tag,
		Component: p.Spec.Component,
	}

	if p.Descriptor == nil {
		if p.Error == nil {
			m.Error = fmt.Sprintf("<unknown component version>")
		} else {
			m.Error = p.Error.Error()
		}
		return m
	}
	norm, hash, err := compdesc.NormHash(p.Descriptor, h.opts.NormAlgorithm, h.opts.Hasher.Create())
	if err != nil {
		m.Error = err.Error()
	} else {
		m.Normalized = string(norm)
		m.Hash = hash
	}
	return m
}

func (h *handler) digester(e interface{}) interface{} {
	p := e.(*comphdlr.Object)
	o := &Object{
		Spec:    p.Spec,
		History: p.History,
	}
	if p.ComponentVersion != nil {
		o.Descriptor = p.ComponentVersion.GetDescriptor()
		if h.mode.action != nil {
			_, o.Descriptor, o.Error = h.mode.action.Digest(p)
		}
	}
	return o
}

func (h *handler) mapGetRegularOutput(e interface{}) interface{} {
	p := e.(*Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	if p.Descriptor == nil {
		if p.Error != nil {
			return []string{p.Spec.Component, tag, "error: " + p.Error.Error()}
		}
		return []string{p.Spec.Component, tag, "<unknown component version>"}
	}
	hash, err := compdesc.Hash(p.Descriptor, h.opts.NormAlgorithm, h.opts.Hasher.Create())
	if err != nil {
		return []string{p.Spec.Component, tag, "error: " + err.Error()}
	}
	return []string{p.Spec.Component, tag, hash}
}

func (h *handler) mapGetWideOutput(e interface{}) interface{} {
	p := e.(*Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	if p.Descriptor == nil {
		return []string{p.Spec.Component, tag, "<unknown component version>"}
	}
	norm, hash, err := compdesc.NormHash(p.Descriptor, h.opts.NormAlgorithm, h.opts.Hasher.Create())
	if err != nil {
		return []string{p.Spec.Component, tag, "error: " + err.Error(), ""}
	}
	return []string{p.Spec.Component, tag, hash, string(norm)}
}
