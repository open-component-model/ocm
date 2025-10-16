package get

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandelsoft/goutils/set"
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm/plugin"
	"ocm.software/ocm/api/ocm/plugin/common"
	utils2 "ocm.software/ocm/api/utils"
	handler "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/pluginhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
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
	loc := p.GetSourceInfo().GetDescription()

	features := p.GetDescriptor().Capabilities()
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

	actions := set.New[string]()
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
