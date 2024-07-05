package set

import (
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/pubsub"
	"github.com/open-component-model/ocm/pkg/out"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

var (
	Names = names.PubSub
	Verb  = verbs.Set
)

type Command struct {
	utils.BaseCommand

	RepoSpec string
	Spec     []byte
}

var _ utils.OCMCommand = (*Command)(nil)

// NewCommand creates a new pubsub command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "{<ocm repository>} [<pub/sub specification>]",
		Short: "Set the pubsub spec for an ocm repository",
		Long: `
A repository may be able to store a publish/subscribe specification
to propagate the creation or update of component version.
If such an implementation is available this command can be used
to set the pub/sub specification for a repository.
If no specification is given an existing specification
will be removed for the given repository.
The specification
can be queried with the <CMD>ocm get pubsub</CMD>.
Types and specification formats are shown for the topic
<CMD>ocm ocm-pubsub</CMD>.
`,
		Args: cobra.RangeArgs(1, 2),
	}
}

func (o *Command) Complete(args []string) error {
	var err error

	o.RepoSpec = args[0]
	if len(args) > 1 {
		o.Spec, err = utils2.ResolveData(args[1], o.FileSystem())
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Command) Run() error {
	var spec cpi.RepositorySpec
	var repo cpi.Repository
	var ps pubsub.PubSubSpec

	u, err := ocm.ParseRepo(o.RepoSpec)
	if err == nil && o.Spec != nil {
		ps, err = pubsub.SpecForData(o, o.Spec)
	}
	if err == nil {
		spec, err = o.OCMContext().MapUniformRepositorySpec(&u)
	}
	if err == nil {
		repo, err = o.OCMContext().RepositoryForSpec(spec)
	}
	if err == nil {
		defer repo.Close()
		if o.Spec == nil {
			ps, err = pubsub.SpecForRepo(repo)
			if err == nil {
				err = pubsub.SetForRepo(repo, nil)
			}
		} else {
			err = pubsub.SetForRepo(repo, ps)
		}
	}
	if err == nil {
		if o.Spec == nil {
			if ps == nil {
				out.Outf(o, "no pubsub spec configured for repository %q\n", o.RepoSpec)
			} else {
				out.Outf(o, "removed pubsub spec %q for repository %q\n", ps.GetKind(), o.RepoSpec)
			}
		} else {
			out.Outf(o, "set pubsub spec %q for repository %q\n", ps.GetKind(), o.RepoSpec)
		}
	}
	return err
}
