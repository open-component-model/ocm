package get

import (
	"strings"

	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/processing"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
)

var (
	Names = names.PubSub
	Verb  = verbs.Get
)

type Command struct {
	utils.BaseCommand

	RepoSpecs []string
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new pubsub command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, output.OutputOptions(outputs))}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "{<ocm repository>}",
		Short: "Get the pubsub spec for an ocm repository",
		Long: `
A repository may be able to store a publish/subscribe specification
to propagate the creation or update of component versions.
If such an implementation is available and a specification is
assigned to the repository, it is shown. The specification
can be set with the <CMD>ocm set pubsub</CMD>.
`,
	}
}

func (o *Command) Complete(args []string) error {
	o.RepoSpecs = args
	return nil
}

func (o *Command) Run() error {
	return utils.HandleOutputsFor("repository spec", output.From(o), o.transform, o.RepoSpecs...)
}

func (o *Command) transform(in string) *Repo {
	var spec cpi.RepositorySpec
	rs := &Repo{RepoSpec: in}
	u, err := ocm.ParseRepo(in)
	if err == nil {
		spec, err = o.OCMContext().MapUniformRepositorySpec(&u)
	}
	if err == nil {
		rs.Repo, err = o.OCMContext().RepositoryForSpec(spec)
	}
	if err == nil {
		rs.Spec, err = pubsub.SpecForRepo(rs.Repo)
	}
	if err != nil {
		rs.Error = err.Error()
	}
	return rs
}

type Repo struct {
	RepoSpec string            `json:"repository"`
	Repo     cpi.Repository    `json:"-"`
	Spec     pubsub.PubSubSpec `json:"pubsub,omitempty"`
	Error    string            `json:"error,omitempty"`
}

var _ output.Manifest = (*Repo)(nil)

func (r *Repo) AsManifest() interface{} {
	return r
}

var outputs = output.NewOutputs(getRegular).AddManifestOutputs()

func TableOutput(opts *output.Options, mapping processing.MappingFunction) *output.TableOutput {
	return &output.TableOutput{
		Headers: output.Fields("REPOSITORY", "PUBSUBTYPE", "ERROR"),
		Options: opts,
		Mapping: mapping,
	}
}

func getRegular(opts *output.Options) output.Output {
	return TableOutput(opts, mapGetRegularOutput).New()
}

func mapGetRegularOutput(e interface{}) interface{} {
	r := e.(*Repo)
	if r.Error != "" {
		return output.Fields(r.RepoSpec, "", r.Error)
	}
	if r.Spec == nil {
		return output.Fields(r.RepoSpec, "-", "")
	}
	list := sliceutils.Slice[string]{}
	Add(r.Repo.GetContext(), r.Spec, &list)
	strings.Join(list, ", ")

	return output.Fields(r.RepoSpec, strings.Join(list, ", "), "")
}

func Add(ctx cpi.Context, s pubsub.PubSubSpec, slice *sliceutils.Slice[string]) {
	if s == nil {
		return
	}
	slice.Add(s.GetKind())
	if u, ok := s.(pubsub.Unwrapable); ok {
		for _, n := range u.Unwrap(ctx) {
			Add(ctx, n, slice)
		}
	}
}
