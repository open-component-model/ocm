package describe

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/spf13/cobra"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	utils3 "ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/names"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/package/bootstrap"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
)

const (
	DEFAULT_CREDENTIALS_FILE = bootstrap.DEFAULT_CREDENTIALS_FILE
	DEFAULT_PARAMETER_FILE   = bootstrap.DEFAULT_PARAMETER_FILE
)

var (
	Names = names.Package
	Verb  = verbs.Describe
)

type Command struct {
	utils.BaseCommand
	Ref string
	Id  metav1.Identity
}

// NewCommand creates a new bootstrap configuration command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), lookupoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] {<component-reference>} {<resource id field>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "describe TOI package",
		Long: `
Describe a TOI package provided by a resource of an OCM component version.

The package resource must have the type <code>` + toi.TypeTOIPackage + `</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic <CMD>ocm toi-bootstrapping</CMD>.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).
`,
		Example: `
$ ocm toi describe package ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
	return cmd
}

func (o *Command) Complete(args []string) error {
	o.Ref = args[0]
	id, err := ocmcommon.MapArgsToIdentityPattern(args[1:]...)
	if err != nil {
		return errors.Wrapf(err, "bootstrap resource identity pattern")
	}
	o.Id = id
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()

	err := o.ProcessOnOptions(ocmcommon.CompleteOptionsWithSession(o, session))
	if err != nil {
		return err
	}
	handler := comphdlr.NewTypeHandler(o.Context.OCM(), session, repooption.From(o).Repository)
	return utils.HandleOutput(&action{cmd: o}, handler, utils.StringElemSpecs(o.Ref)...)
}

////////////////////////////////////////////////////////////////////////////////

type action struct {
	data comphdlr.Objects
	cmd  *Command
}

var _ output.Output = (*action)(nil)

func (a *action) Add(e interface{}) error {
	o, ok := e.(*comphdlr.Object)
	if !ok {
		return fmt.Errorf("object of type %T is not a valid comphdlr.Object", e)
	}
	if o.ComponentVersion != nil && !ocireg.IsKind(o.Repository.GetSpecification().GetKind()) {
		if ociuploadattr.Get(a.cmd.Context) == nil {
			out.Outf(a.cmd, "Warning: repository is no OCI registry, consider importing it or use upload repository with option ' -X ociuploadrepo=...\n")
		} else {
			out.Outf(a.cmd, "Warning: repository is no OCI registry, consider importing.\n'")
		}
	}
	a.data = append(a.data, o)
	return nil
}

func (a *action) Close() error {
	return nil
}

type Binary struct {
	Binary []byte `json:"binary"`
}

func (a *action) Out() error {
	cnt := 0
	for i := range a.data {
		if i > 0 {
			a.Outf("\n")
		}
		nv := common.VersionedElementKey(a.data[i].ComponentVersion)
		err := a.describe(a.data[i].ComponentVersion)
		if err != nil {
			out.Errf(a.cmd.Context, "%s: %s\n", nv, err)
			cnt++
		}
	}
	if cnt > 0 {
		return fmt.Errorf("describe failed for %d packages", cnt)
	}
	return nil
}

func (a *action) Outf(msg string, args ...interface{}) {
	out.Outf(a.cmd.Context, msg, args...)
}

type einfo struct {
	index int
	ectx  *install.ExecutorContext
}

func (a *action) describe(cv ocm.ComponentVersionAccess) error {
	nv := common.VersionedElementKey(cv)
	rid := metav1.NewResourceRef(a.cmd.Id)
	resolver := lookupoption.From(a.cmd)

	ires, eff, err := resourcerefs.MatchResourceReference(cv, toi.TypeTOIPackage, rid, resolver)
	if err != nil {
		return errors.Wrapf(err, "package resource in %s", nv)
	}
	defer eff.Close()

	var spec toi.PackageSpecification
	err = install.GetResource(ires, &spec)
	if err != nil {
		return errors.ErrInvalidWrap(err, "package spec")
	}

	a.Outf("TOI Package %s[%s]\n", nv, ires.Meta().GetName())

	if spec.Description != "" {
		a.Outf("  Package Description:\n%s\n\n", utils3.IndentLines(strings.TrimSpace(spec.Description), "    ", false))
	}
	if len(spec.AdditionalResources) == 0 {
		a.Outf("  no additional resources found\n")
	} else {
		keys := utils3.StringMapKeys(spec.AdditionalResources)
		a.Outf("  Additional Resources:\n")
		for _, k := range keys {
			switch k {
			case toi.AdditionalResourceCredentialsFile:
				out.Outf(a.cmd.Context, "    - %s: %s\n", k, "downloadable credential file template)")
			case toi.AdditionalResourceConfigFile:
				out.Outf(a.cmd.Context, "    - %s: %s\n", k, "downloadable user configuration file template)")
			default:
				out.Outf(a.cmd.Context, "    - %s\n", k)
			}
		}
	}

	actions := map[string]*einfo{}

	for i, e := range spec.Executors {
		eacts := e.Actions

		ectx, err := install.DetermineExecutor(&e, a.cmd.OCMContext(), cv, resolver)
		if err != nil {
			a.Outf("  Warning: cannot determine executor %s: %s\n", e.Name(), err)
		}
		info := &einfo{i, ectx}
		if len(eacts) == 0 && ectx != nil {
			eacts = ectx.Spec.Actions
		}
		if len(eacts) == 0 {
			eacts = []string{"<default>"}
		}
		for o, n := range eacts {
			if _, ok := actions[n]; ok {
				a.Outf("  Warning: action %s defined for multiple executors: %d and %d\n", n, o, i)
			} else {
				actions[n] = info
			}
		}
	}
	if len(actions) == 0 {
		a.Outf("  Warning: no actions defined\n")
	} else {
		keys := utils3.StringMapKeys(actions)
		a.Outf("  Supported Actions:\n")
		for _, k := range keys {
			info := actions[k]
			exec := spec.Executors[info.index]
			out.Outf(a.cmd.Context, "    - %s: provided by %s\n", k, exec.Name())
			if len(exec.CredentialMapping) > 0 {
				ckeys := utils3.StringMapKeys(exec.CredentialMapping)
				out.Outf(a.cmd.Context, "      credential key mappings\n")
				for _, c := range ckeys {
					out.Outf(a.cmd.Context, "      - %s: %s\n", c, exec.CredentialMapping[c])
				}
			}
		}
	}
	if len(spec.Credentials) == 0 {
		a.Outf("  no credentials required\n")
	} else {
		keys := utils3.StringMapKeys(spec.Credentials)
		a.Outf("  Required Credentials:\n")
		for _, k := range keys {
			cred := spec.Credentials[k]
			opt := ""
			if cred.Optional {
				opt = " (optional)"
			}
			out.Outf(a.cmd.Context, "    - %s%s\n", k, opt)
			out.Outf(a.cmd.Context, "      description: %s\n", utils3.IndentLines(strings.TrimSpace(cred.Description), "      ", true))
			if len(cred.ConsumerId) != 0 {
				out.Outf(a.cmd.Context, "      used as consumer id: %s\n", cred.ConsumerId)
			}
			if len(cred.Properties) != 0 {
				ckeys := utils3.StringMapKeys(cred.Properties)
				out.Outf(a.cmd.Context, "      required properties:\n")
				for _, c := range ckeys {
					out.Outf(a.cmd.Context, "      - %s: %s\n", c, utils3.IndentLines(strings.TrimSpace(cred.Properties[c]), "        ", true))
				}
			}
		}
	}
	return nil
}
