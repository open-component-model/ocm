package list

import (
	"fmt"

	"github.com/spf13/cobra"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/vershdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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
all versions according to the given versuin constraints are listed.
`,
		Example: `
$ ocm list componentversion ghcr.io/mandelsoft/kubelink
$ ocm list componentversion --repo OCIRegistry::ghcr.io mandelsoft/kubelink
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
