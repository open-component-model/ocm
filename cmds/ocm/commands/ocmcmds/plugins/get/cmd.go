// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	handler "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/common"
	"github.com/open-component-model/ocm/pkg/generics"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

var (
	Names = names.Plugins
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Names []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx, output.OutputOptions(outputs)),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<plugin name>}",
		Short: "get plugins",
		Long: `
Get lists information for all plugins specified, if no plugin is specified
all registered ones are listed.
`,
		Example: `
$ ocm get plugins
$ ocm get plugins demo -o yaml
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.Names = args
	return nil
}

func (o *Command) Run() error {
	hdlr := handler.NewTypeHandler(o.Context.OCM())
	return utils.HandleArgs(output.From(o), hdlr, o.Names...)
}

/////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("PLUGIN", "VERSION", "SOURCE", "DESCRIPTION", wide),
		Options: opts,
		Mapping: mapping,
	}
	return def
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetRegularOutput, "CAPABILITIES").New()
}

func getWide(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetWideOutput, "ACCESSMETHODS", "UPLOADERS", "DOWNLOADERS", "ACTIONS").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	loc := "local"
	src := p.GetSource()
	if src != nil {
		loc = src.Component + ":" + src.Version
	}

	var features []string
	if len(p.GetDescriptor().AccessMethods) > 0 {
		features = append(features, "accessmethods")
	}
	if len(p.GetDescriptor().Uploaders) > 0 {
		features = append(features, "uploaders")
	}
	if len(p.GetDescriptor().Downloaders) > 0 {
		features = append(features, "downloaders")
	}
	if len(p.GetDescriptor().Actions) > 0 {
		features = append(features, "actions")
	}
	if len(p.GetDescriptor().ValueSets) > 0 {
		features = append(features, "valuesets")
	}
	if len(p.GetDescriptor().ValueMergeHandlers) > 0 {
		features = append(features, "mergehandlers")
	}
	if len(p.GetDescriptor().LabelMergeSpecifications) > 0 {
		features = append(features, "mergespecs")
	}

	return []string{p.Name(), p.Version(), loc, p.Message(), strings.Join(features, ",")}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	d := p.GetDescriptor()

	found := map[string][]string{}
	for _, m := range d.AccessMethods {
		l := found[m.Name]
		v := m.Version
		if v != "" {
			l = append(l, v)
		}
		found[m.Name] = l
	}

	var methods []string
	for _, m := range utils2.StringMapKeys(found) {
		l := found[m]
		if len(l) == 0 {
			methods = append(methods, m)
		} else {
			sort.Strings(l)
			methods = append(methods, fmt.Sprintf("%s[%s]", m, strings.Join(l, ",")))
		}
	}

	actions := generics.Set[string]{}
	for _, a := range d.Actions {
		actions.Add(a.Name)
	}
	actionList := actions.AsArray()
	sort.Strings(actionList)

	// a working type inference would be really great
	ups := common.DescribeElements[plugin.UploaderDescriptor, plugin.UploaderKey](d.Uploaders)
	downs := common.DescribeElements[plugin.DownloaderDescriptor, plugin.DownloaderKey](d.Downloaders)

	reg := output.Fields(mapGetRegularOutput(e))
	return output.Fields(reg[:len(reg)-1], strings.Join(methods, ","), ups, downs, strings.Join(actionList, ","))
}
