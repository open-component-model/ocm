package get

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	compdescv2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/schemaoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/versionconstraintsoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Components
	Verb  = verbs.Get
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
				closureoption.New("component reference", output.Fields("IDENTITY"), options.Not(output.Selected("tree")), addIdentityField),
				lookupoption.New(),
				schemaoption.New("", true),
			))},
		utils.Names(Names, names...)...,
	)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<component-reference>}",
		Short: "get component version",
		Long: `
Get lists all component versions specified, if only a component is specified
all versions are listed.
`,
		Example: `
$ ocm get componentversion ghcr.io/open-component-model/ocm//ocm.software/ocmcli:0.17.0
$ ocm get componentversion --repo OCIRegistry::ghcr.io/open-component-model/ocm ocm.software/ocmcli:0.17.0
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

func (o *Command) Run() (err error) {
	session := ocm.NewSession(nil)
	defer errors.PropagateError(&err, session.Close)

	err = o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository, comphdlr.OptionsFor(o))
	return utils.HandleArgs(output.From(o).WithSession(session), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func addIdentityField(e interface{}) []string {
	p := e.(*comphdlr.Object)
	return []string{p.Identity.String()}
}

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	def := &output.TableOutput{
		Headers: output.Fields("COMPONENT", "VERSION", "PROVIDER", wide),
		Options: opts,
		Chain:   comphdlr.Sort,
		Mapping: mapping,
	}
	return closureoption.TableOutput(def, comphdlr.ClosureExplode)
}

/////////////////////////////////////////////////////////////////////////////

type FailedEntry struct {
	Scheme  string `json:"scheme,omitempty"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Error   string `json:"error"`
}

func Format(opts *output.Options) processing.ProcessChain {
	o := schemaoption.From(opts)
	if o.Schema == compdesc.InternalSchemaVersion {
		return nil
	}
	return processing.Map(func(in interface{}) interface{} {
		cv := comphdlr.Elem(in)
		if cv == nil {
			nv := in.(*comphdlr.Object).Spec.NameVersion()
			return &FailedEntry{
				Name:    nv.GetName(),
				Version: nv.GetVersion(),
				Error:   "not found",
			}
		}
		desc := cv.GetDescriptor()
		schema := o.Schema
		if schema == "" {
			schema = desc.SchemaVersion()
		}
		if schema == "" {
			schema = compdescv2.SchemaVersion
		}
		out, err := compdesc.Convert(desc, compdesc.SchemaVersion(schema))
		if err != nil {
			return &FailedEntry{
				Scheme:  desc.SchemaVersion(),
				Name:    desc.GetName(),
				Version: desc.GetVersion(),
				Error:   err.Error(),
			}
		} else {
			return out
		}
	})
}

/////////////////////////////////////////////////////////////////////////////

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
	"tree": getTree,
}).AddChainedManifestOutputs(output.ComposeChain(closureoption.OutputChainFunction(comphdlr.ClosureExplode, comphdlr.Sort), Format))

func getRegular(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetRegularOutput).New()
}

func getWide(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetWideOutput, "REPOSITORY").New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetRegularOutput), "NESTING").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	p := e.(*comphdlr.Object)

	tag := "-"
	if p.Spec.Version != nil {
		tag = *p.Spec.Version
	}
	if p.ComponentVersion == nil {
		return []string{p.Spec.Component, tag, "<unknown component version>"}
	}
	return []string{p.Spec.Component, tag, string(p.ComponentVersion.GetDescriptor().Provider.Name)}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := e.(*comphdlr.Object)
	return output.Fields(mapGetRegularOutput(e), p.Spec.UniformRepositorySpec.String())
}
