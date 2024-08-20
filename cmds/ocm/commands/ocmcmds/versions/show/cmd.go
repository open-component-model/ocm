package show

import (
	"github.com/Masterminds/semver/v3"
	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/utils/semverutils"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

var (
	Names = names.Versions
	Verb  = verbs.Show
)

type Command struct {
	utils.BaseCommand
	Latest   bool
	Semantic bool

	Ref         string
	Constraints []*semver.Constraints
}

// NewCommand creates a new ocm command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx,
		repooption.New(),
	)}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component> {<version pattern>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "show dedicated versions (semver compliant)",
		Long: `
Match versions of a component against some patterns.
`,
		Example: `
$ ocm show versions ghcr.io/mandelsoft/cnudie//github.com/mandelsoft/playground
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.BoolVarP(&o.Latest, "latest", "l", false, "show only latest version")
	fs.BoolVarP(&o.Semantic, "semantic", "s", false, "show semantic version")
}

func (o *Command) Complete(args []string) error {
	o.Ref = args[0]

	for _, v := range args[1:] {
		c, err := semver.NewConstraint(v)
		if err != nil {
			return err
		}
		o.Constraints = append(o.Constraints, c)
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

	versions := Versions{}
	repo := repooption.From(o)

	var cv ocm.ComponentVersionAccess
	var comp ocm.ComponentAccess

	// determine version source
	if repo.Repository != nil {
		cr, err := ocm.ParseComp(o.Ref)
		if err != nil {
			return err
		}
		comp, err = session.LookupComponent(repo.Repository, cr.Component)
		if err != nil {
			return err
		}
		if cr.IsVersion() {
			cv, err = session.GetComponentVersion(comp, *cr.Version)
			if err != nil {
				return err
			}
		}
	} else {
		r, err := session.EvaluateVersionRef(o.Context.OCMContext(), o.Ref)
		if err != nil {
			return err
		}
		if r.Component == nil {
			return errors.Newf("no component specified")
		}
		comp = r.Component
		cv = r.Version
	}

	// determine version base set
	if cv != nil {
		v, err := semver.NewVersion(cv.GetVersion())
		if err != nil {
			return err
		}
		versions = append(versions, v)
	} else {
		vers, err := comp.ListVersions()
		if err != nil {
			return err
		}
		for _, vn := range vers {
			v, err := semver.NewVersion(vn)
			if err == nil {
				versions = append(versions, v)
			}
		}
	}

	versions = semverutils.MatchVersions(versions, o.Constraints...)
	if len(versions) > 1 && o.Latest {
		versions = versions[len(versions)-1:]
	}

	for _, r := range versions {
		if o.Semantic {
			out.Outf(o, "%s\n", r)
		} else {
			out.Outf(o, "%s\n", r.Original())
		}
	}
	return nil
}

type Versions = semver.Collection

/*
var _ sort.Interface = (Versions)(nil)

func (v Versions) Len() int {
	return len(v)
}

func (v Versions) Less(i, j int) bool {
	return v[i].Compare(v[j])<0
}

func (v Versions) Swap(i, j int) {
	v[i],v[j]=v[j],v[i]
}

*/
