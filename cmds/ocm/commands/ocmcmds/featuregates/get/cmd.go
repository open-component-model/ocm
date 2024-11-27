package get

import (
	"encoding/json"

	"github.com/mandelsoft/goutils/general"
	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/datacontext/attrs/featuregatesattr"
	"ocm.software/ocm/api/ocm/extensions/featuregates"
	handler "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/featurehdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.FeatureGates
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Names       []string
	OutputMode  string
	MatcherType string

	Matcher  credentials.IdentityMatcher
	Consumer credentials.ConsumerIdentity
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
		Use:   "[<options>] {<name>}",
		Short: "list feature gates",
		Args:  cobra.MinimumNArgs(0),
		Long: `
Show feature gates and the activation.

The following feature gates are supported:
` + featuregates.Usage(featuregates.DefaultRegistry()),
		Example: `
$ ocm get featuregates
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Complete(args []string) error {
	o.Names = args
	return nil
}

func (o *Command) Run() error {
	hdlr := handler.NewTypeHandler(o.Context.OCMContext())
	return utils.HandleArgs(output.From(o), hdlr, o.Names...)
}

////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("FEATURE", "ENABLED", "MODE", "DESCRIPTION", wide),
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
	return TableOutput(opts, mapGetWideOutput, "ATTRIBUTES").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := handler.Elem(e)

	return []string{p.Name, general.Conditional(p.Mode == featuregatesattr.FEATURE_DISABLED, "disabled", "enabled"), p.Mode, p.Short}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := handler.Elem(e)

	attr := ""
	if len(p.Attributes) > 0 {
		data, err := json.Marshal(p.Attributes)
		if err == nil {
			attr = string(data)
		} else {
			attr = err.Error()
		}
	}
	reg := output.Fields(mapGetRegularOutput(e))
	return output.Fields(reg, attr)
}
