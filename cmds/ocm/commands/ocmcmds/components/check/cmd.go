// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package check

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/errors"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

var (
	Names = names.Components
	Verb  = verbs.Check
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx,
				repooption.New(),
				output.OutputOptions(outputs),
			),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "check completeness of a component version in an OCM repository",
		Long: `
THis command checks, whether component versiuons are completely contained
in an OCM repository with all its dependent component references.
`,
		Example: `
$ ocm check componentversion ghcr.io/mandelsoft/kubelink
$ ocm get componentversion --repo OCIRegistry::ghcr.io mandelsoft/kubelink
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
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, comphdlr.OptionsFor(o))
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

////////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(OutputFactory(mapRegularOutput), output.Outputs{
	"wide": OutputFactory(mapWideOutput, "MISSING"),
}).AddChainedManifestOutputs(CreateChain)

func OutputFactory(fmt processing.MappingFunction, wide ...string) output.OutputFactory {
	return func(opts *output.Options) output.Output {
		return (&output.TableOutput{
			Headers: output.Fields("COMPONENT", "VERSION", "STATUS", "ERROR", wide),
			Options: opts,
			Chain:   NewAction(),
			Mapping: fmt,
		}).New()
	}
}

func mapRegularOutput(e interface{}) interface{} {
	p := e.(*Entry)

	err := ""
	if p.Error != nil {
		err = p.Error.Error()
	}
	if len(p.Missing) == 0 {
		return []string{p.ComponentVersion.GetName(), p.ComponentVersion.GetVersion(), p.Status, err}
	}
	return []string{p.ComponentVersion.GetName(), p.ComponentVersion.GetVersion(), p.Status, err}
}

func mapWideOutput(e interface{}) interface{} {
	p := e.(*Entry)

	err := ""
	if p.Error != nil {
		err = p.Error.Error()
	}
	if len(p.Missing) == 0 {
		return []string{p.ComponentVersion.GetName(), p.ComponentVersion.GetVersion(), p.Status, err, ""}
	}
	missing := map[string]string{}
	for id, m := range p.Missing {
		sep := "["
		d := ""
		for _, id := range m[:len(m)-1] {
			d = d + sep + id.String()
			sep = "->"
		}
		missing[id.String()] = d + "]"
	}
	msg := ""
	sep := ""
	for _, k := range utils2.StringMapKeys(missing) {
		msg += sep + k + missing[k]
		sep = ", "
	}
	return []string{p.ComponentVersion.GetName(), p.ComponentVersion.GetVersion(), p.Status, err, msg}
}

////////////////////////////////////////////////////////////////////////////////

type Missing map[common.NameVersion]common.History
type Entry struct {
	Status           string             `json:"status"`
	ComponentVersion common.NameVersion `json:"componentVersion"`
	Missing          Missing            `json:"missing,omitempty"`
	Error            error              `json:"error,omitempty"`
}

func (n Missing) MarshalJSON() ([]byte, error) {
	m := map[string]common.History{}
	for k, v := range n {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func CreateChain(options *output.Options) processing.ProcessChain {
	// opts are not required by action so far.
	return NewAction()
}

type action struct {
	cache map[common.NameVersion]map[common.NameVersion]common.History
}

func NewAction() processing.ProcessChain {
	return comphdlr.Sort.Map((&action{}).Map)
}

func (a *action) Map(in interface{}) interface{} {
	a.cache = map[common.NameVersion]map[common.NameVersion]common.History{}

	i := in.(*comphdlr.Object)
	o := &Entry{
		Status:           "OK",
		ComponentVersion: common.VersionedElementKey(i.ComponentVersion),
	}
	o.Missing, o.Error = a.handle(i.ComponentVersion, common.History{common.VersionedElementKey(i.ComponentVersion)})
	if o.Error != nil {
		o.Status = "Error"
	}
	if len(o.Missing) > 0 {
		o.Status = "Incomplete"
	}
	return o
}

func (a *action) getMissing(repo ocm.Repository, id common.NameVersion, h common.History) (Missing, error) {
	if r, ok := a.cache[id]; ok {
		return r, nil
	}

	err := h.Add(ocm.KIND_COMPONENTVERSION, id)
	if err != nil {
		return nil, err
	}
	cv, err := repo.LookupComponentVersion(id.GetName(), id.GetVersion())
	if err != nil {
		if !errors.IsErrNotFound(err) {
			return nil, err
		}
	}
	if cv == nil {
		return map[common.NameVersion]common.History{id: h}, nil
	} else {
		return a.handle(cv, h)
	}
}

func (a *action) handle(cv ocm.ComponentVersionAccess, h common.History) (Missing, error) {
	var missing Missing
	for _, r := range cv.GetDescriptor().References {
		id := common.NewNameVersion(r.ComponentName, r.Version)
		n, err := a.getMissing(cv.Repository(), id, h)
		if err != nil {
			return missing, err
		}
		if len(n) > 0 {
			if missing == nil {
				missing = Missing{}
			}
			for k, v := range n {
				missing[k] = v
			}
		}
	}
	return missing, nil
}
