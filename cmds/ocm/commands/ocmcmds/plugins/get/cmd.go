// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	handler "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin"
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
			BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs)),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<plugin name>}",
		Short: "get plugins",
		Example: `
$ ocm get plugins
$ ocm get plugins demo -o yaml
`,
	}
}

func (o *Command) Long() string {
	return `
Get lists information for all plugins specified, if no plugin is specified
all registered ones are listed.
`
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
		Headers: output.Fields("PLUGIN", "VERSION", "DESCRIPTION", wide),
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
	return TableOutput(opts, mapGetRegularOutput).New()
}

func getWide(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetWideOutput, "ACCESSMETHODS", "UPLOADERS", "DOWNLOADERS").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	return []string{p.Name(), p.Version(), p.Message()}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	d := p.GetDescriptor()

	var list []string
	for _, m := range d.AccessMethods {
		n := m.Name
		if m.Version != "" {
			n += "/" + m.Version
		}
		list = append(list, n)
	}

	// a working type inference would be really great
	ups := Describe[plugin.UploaderDescriptor, plugin.UploaderKey](d.Uploaders)
	downs := Describe[plugin.DownloaderDescriptor, plugin.DownloaderKey](d.Downloaders)

	return output.Fields(mapGetRegularOutput(e), strings.Join(list, ","), ups, downs)
}

func Describe[E Element[C], C Stringer](elems []E) string {
	var list []string
	for _, m := range elems {
		n := m.GetName()
		var clist []string
		for _, c := range m.GetConstraints() {
			clist = append(clist, c.String())
		}
		if len(clist) > 0 {
			n = fmt.Sprintf("%s[%s]", n, strings.Join(clist, ","))
		}
		list = append(list, n)
	}
	return strings.Join(list, ",")
}

type Stringer interface {
	String() string
}

type Element[C Stringer] interface {
	GetName() string
	GetConstraints() []C
}
