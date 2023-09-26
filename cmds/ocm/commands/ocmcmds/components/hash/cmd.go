// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package hash

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"
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
$ ocm hash componentversion --repo OCIRegistry::ghcr.io mandelsoft/kubelink
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

func TableOutput(opts *output.Options, h *action, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
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
	"norm": getNorm,
}).AddChainedManifestOutputs(output.ComposeChain(closureoption.OutputChainFunction(comphdlr.ClosureExplode, comphdlr.Sort), mapManifest))

func mapManifest(opts *output.Options) processing.ProcessChain {
	h := newAction(opts)
	return processing.Map(h.digester).Map(h.manifester)
}

func getRegular(opts *output.Options) output.Output {
	h := newAction(opts)
	return TableOutput(opts, h, h.mapGetRegularOutput).New()
}

func getWide(opts *output.Options) output.Output {
	h := newAction(opts)
	return TableOutput(opts, h, h.mapGetWideOutput, "NORMALIZED FORM").New()
}

func getNorm(opts *output.Options) output.Output {
	h := newAction(opts)
	return h
}

////////////////////////////////////////////////////////////////////////////////

// action as output.Output.
type action struct {
	ctx  clictx.Context
	opts *hashoption.Option
	mode *Option

	norms map[common.NameVersion]string
}

func (h *action) Add(e interface{}) error {
	m := h._manifester(h.digester(e))
	if m.Error != "" {
		return fmt.Errorf("cannot handle %s: %s\n", m.History, m.Error)
	}
	h.norms[common.NewNameVersion(m.Component, m.Version)] = m.Normalized
	return nil
}

func (h *action) Close() error {
	return nil
}

func (h *action) Out() error {
	if len(h.norms) > 1 {
		dir := h.mode.outfile
		if strings.HasSuffix(dir, ".ncd") {
			dir = dir[:len(dir)-4]
		}
		err := h.ctx.FileSystem().Mkdir(dir, 0o755)
		if err != nil {
			return fmt.Errorf("cannot create output dir %s", dir)
		}
		for k, n := range h.norms {
			p := filepath.Join(dir, k.String())
			err := h.write(p+".ncd", n)
			if err != nil {
				return err
			}
		}
	} else {
		for _, n := range h.norms {
			return h.write(h.mode.outfile, n)
		}
	}
	return nil
}

func (h *action) write(p, n string) error {
	dir := filepath.Dir(p)
	err := h.ctx.FileSystem().MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf("cannot create dir %s", dir)
	}
	return vfs.WriteFile(h.ctx.FileSystem(), p, []byte(n), 0o644)
}

/////////

func newAction(opts *output.Options) *action {
	h := &action{
		ctx:  opts.Context,
		opts: hashoption.From(opts),
		mode: From(opts),
	}
	if opts.OutputMode == "norm" {
		h.norms = map[common.NameVersion]string{}
	}
	return h
}

func (h *action) manifester(e interface{}) interface{} {
	return h._manifester(e)
}

func (h *action) _manifester(e interface{}) *Manifest {
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

func (h *action) digester(e interface{}) interface{} {
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

func (h *action) mapGetRegularOutput(e interface{}) interface{} {
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

func (h *action) mapGetWideOutput(e interface{}) interface{} {
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
