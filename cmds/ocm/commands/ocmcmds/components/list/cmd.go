package list

import (
	"fmt"

	"github.com/spf13/cobra"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/vershdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Components
	Verb  = verbs.List
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(
		&Command{BaseCommand: utils.NewBaseCommand(ctx,
			versionconstraintsoption.New(), repooption.New(),
			output.OutputOptions(outputs,
				lookupoption.New(),
				schemaoption.New("", true),
			))},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "list component version names",
		Long: `
List lists the version names of the specified objects, if only a component is specified
all versions according to the given version constraints are listed.
`,
		Example: `
$ ocm list componentversion ghcr.io/open-component-model/ocm//ocm.software/ocmcli
$ ocm list componentversion --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
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
	handler := vershdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, vershdlr.OptionsFor(o))
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func TableOutput(opts *output.Options, mapping processing.MappingFunction) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", "MESSAGE"),
		Options: opts,
		Chain:   vershdlr.Sort,
		Mapping: mapping,
	}
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{}).AddManifestOutputs()

func getRegular(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetRegularOutput).New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := vershdlr.Elem(e)

	tag := ""
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	if p.Version == "" {
		return []string{p.Component, tag, "<unknown component version>"}
	}
	return []string{p.Spec.Component, tag, ""}
}
