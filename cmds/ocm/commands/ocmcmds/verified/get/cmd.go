package get

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	handler "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/verifiedhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/storeoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Verified
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	path  string
	Names []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{
			BaseCommand: utils.NewBaseCommand(ctx, output.OutputOptions(outputs).SortColumns("VERSION", "COMPONENT")),
		},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component / version}",
		Short: "get verified component versions",
		Long: `
Get lists remembered verified component versions.
`,
		Example: `
$ ocm get verified
$ ocm get verified -f verified.yaml acme.org/component -o yaml
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.path, "verified", "", storeoption.DEFAULT_VERIFIED_FILE, "verified file")
}

func (o *Command) Complete(args []string) error {
	o.Names = args
	return nil
}

func (o *Command) Run() error {
	hdlr, err := handler.NewTypeHandler(o.Context.OCM(), o.path)
	if err != nil {
		return err
	}
	return utils.HandleArgs(output.From(o), hdlr, o.Names...)
}

/////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", wide),
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
	return TableOutput(opts, mapGetWideOutput, "SIGNATURES").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	return []string{p.Element.GetName(), p.Element.GetVersion()}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := handler.Elem(e)
	return []string{p.Element.GetName(), p.Element.GetVersion(), strings.Join(p.Signatures, ", ")}
}
