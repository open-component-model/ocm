package get

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/elemhdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/sources/common"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Sources
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Comp string
	Ids  []metav1.Identity
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, versionconstraintsoption.New(), repooption.New(), output.OutputOptions(outputs, closureoption.New("component reference"), lookupoption.New()))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <component> {<name> { <key>=<value> }}",
		Args:  cobra.MinimumNArgs(1),
		Short: "get sources of a component version",
		Long: `
Get sources of a component version. Sources are specified
by identities. An identity consists of 
a name argument followed by optional <code>&lt;key>=&lt;value></code>
arguments.
`,
	}
}

func (o *Command) Complete(args []string) error {
	var err error
	o.Comp = args[0]
	o.Ids, err = ocmcommon.MapArgsToIdentities(args[1:]...)
	return err
}

func (o *Command) Run() (err error) {
	session := ocm.NewSession(nil)
	defer errors.PropagateError(&err, session.Close)

	err = o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}

	opts := output.From(o)
	hdlr, err := common.NewTypeHandler(o.Context.OCM(), opts, repooption.From(o).Repository, session, []string{o.Comp}, common.OptionsFor(o))
	if err != nil {
		return err
	}
	specs, err := utils.ElemSpecs(o.Ids)
	if err != nil {
		return err
	}

	return utils.HandleOutputs(opts, hdlr, specs...)
}

////////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields(elemhdlr.MetaOutput, "TYPE", wide),
		Options: opts,
		Chain:   elemhdlr.Sort,
		Mapping: mapping,
	}
}

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
	"tree": getTree,
}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetRegularOutput)).New()
}

func getWide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetWideOutput, elemhdlr.AccessOutput...)).New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetRegularOutput), "COMPONENT", elemhdlr.NodeMapping(1), output.TreeSynthesizedTitleFunc(elemhdlr.NodeTitle)).New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	r := common.Elem(e)
	return sliceutils.CopyAppend(elemhdlr.MapMetaOutput(e), r.Type)
}

func mapGetWideOutput(e interface{}) interface{} {
	return sliceutils.CopyAppend(mapGetRegularOutput(e).([]string), elemhdlr.MapAccessOutput(common.Elem(e).Access)...)
}
