package get

import (
	"fmt"

	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/cmds/ocm/commands/common/options/closureoption"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/common/handlers/artifacthdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocicmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/options"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/processing"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Artifacts
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	Refs []string
}

// NewCommand creates a new artifact command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), output.OutputOptions(outputs, &Attached{}, closureoption.New("index")))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] {<artifact-reference>}",
		Short: "get artifact version",
		Long: `
Get lists all artifact versions specified, if only a repository is specified
all tagged artifacts are listed.
	`,
		Example: `
$ ocm get artifact ghcr.io/open-component-model/ocm/component-descriptors/ocm.software/ocmcli
$ ocm get artifact ghcr.io/open-component-model/ocm/ocm.software/ocmcli/ocmcli-image:0.17.0
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) Complete(args []string) error {
	if len(args) == 0 && repooption.From(o).Spec == "" {
		return fmt.Errorf("a repository or at least one argument that defines the reference is needed")
	}
	o.Refs = args
	return nil
}

func (o *Command) Run() error {
	session := oci.NewSession(nil)
	defer session.Close()
	err := o.ProcessOnOptions(common.CompleteOptionsWithContext(o.Context, session))
	if err != nil {
		return err
	}
	handler := artifacthdlr.NewTypeHandler(o.Context.OCI(), session, repooption.From(o).Repository)
	return utils.HandleArgs(output.From(o), handler, o.Refs...)
}

/////////////////////////////////////////////////////////////////////////////

func OutputChainFunction() output.ChainFunction {
	return func(opts *output.Options) processing.ProcessChain {
		chain := closureoption.Closure(opts, artifacthdlr.ClosureExplode, artifacthdlr.Sort)
		chain = processing.Append(chain, artifacthdlr.ExplodeAttached, AttachedFrom(opts))
		return processing.Append(chain, artifacthdlr.Clean, options.Or(closureoption.From(opts), output.OutputModeCondition(opts, "tree")))
	}
}

func TableOutput(opts *output.Options, mapping processing.MappingFunction, wide ...string) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields("REGISTRY", "REPOSITORY", "KIND", "TAG", "DIGEST", wide),
		Chain:   OutputChainFunction()(opts),
		Options: opts,
		Mapping: mapping,
	}
}

var outputs = output.NewOutputs(getRegular, output.Outputs{
	"wide": getWide,
	"tree": getTree,
}).AddChainedManifestOutputs(OutputChainFunction())

func getRegular(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetRegularOutput)).New()
}

func getWide(opts *output.Options) output.Output {
	return closureoption.TableOutput(TableOutput(opts, mapGetWideOutput, "MIMETYPE", "CONFIGTYPE")).New()
}

func getTree(opts *output.Options) output.Output {
	return output.TreeOutput(TableOutput(opts, mapGetRegularOutput), "NESTING").New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	digest := "unknown"
	p := e.(*artifacthdlr.Object)
	blob, err := p.Artifact.Blob()
	if err == nil {
		digest = blob.Digest().String()
	}
	tag := "-"
	if p.Spec.Tag != nil {
		tag = *p.Spec.Tag
	}
	kind := "-"
	if p.Artifact.IsManifest() {
		kind = "manifest"
	}
	if p.Artifact.IsIndex() {
		kind = "index"
	}
	return []string{p.Spec.UniformRepositorySpec.String(), p.Spec.Repository, kind, tag, digest}
}

func mapGetWideOutput(e interface{}) interface{} {
	p := e.(*artifacthdlr.Object)

	config := "-"
	if p.Artifact.IsManifest() {
		config = p.Artifact.ManifestAccess().GetDescriptor().Config.MediaType
	}
	return output.Fields(mapGetRegularOutput(e), p.Artifact.GetDescriptor().MimeType(), config)
}
