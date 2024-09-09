package bootstrap

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/resourcerefs"
	"ocm.software/ocm/api/ocm/tools/toi"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	utils3 "ocm.software/ocm/api/utils"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/utils/runtime"
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
	Names = names.Configuration
	Verb  = verbs.Bootstrap
)

type Command struct {
	utils.BaseCommand
	Ref string
	Id  metav1.Identity

	CredentialsFile string
	ParameterFile   string
}

// NewCommand creates a new bootstrap configuration command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), lookupoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] {<component-reference>} {<resource id field>}",
		Args:  cobra.MinimumNArgs(1),
		Short: "bootstrap TOI configuration files",
		Long: `
If a TOI package provides information for configuration file templates/prototypes
this command extracts this data and provides appropriate files in the filesystem.

The package resource must have the type <code>` + toi.TypeTOIPackage + `</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic <CMD>ocm toi-bootstrapping</CMD>.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no credentials file name is provided (option -c) the file
<code>` + DEFAULT_CREDENTIALS_FILE + `</code> is used. If no parameter file name is
provided (option -p) the file <code>` + DEFAULT_PARAMETER_FILE + `</code> is used.

For more details about those files see <CMD>ocm bootstrap package</CMD>.
`,
		Example: `
$ ocm toi bootstrap config ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringVarP(&o.CredentialsFile, "credentials", "c", DEFAULT_CREDENTIALS_FILE, "credentials file name")
	fs.StringVarP(&o.ParameterFile, "parameters", "p", DEFAULT_PARAMETER_FILE, "parameter file name")
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
	if len(a.data) > 0 {
		return errors.New("found multiple component versions")
	}
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
	cv := a.data[0].ComponentVersion
	nv := common.VersionedElementKey(cv)
	rid := metav1.NewResourceRef(a.cmd.Id)
	resolver := lookupoption.From(a.cmd)

	ires, eff, err := resourcerefs.MatchResourceReference(cv, toi.TypeTOIPackage, rid, resolver)
	if err != nil {
		return errors.Wrapf(err, "package resource in %s", nv)
	}
	defer eff.Close()
	out.Outf(a.cmd.Context, "found package resource %q in %s\n", ires.Meta().GetName(), nv)

	var spec toi.PackageSpecification
	err = install.GetResource(ires, &spec)
	if err != nil {
		return errors.ErrInvalidWrap(err, "package spec")
	}

	if spec.Description != "" {
		out.Outf(a.cmd.Context, "\nPackage Description:\n%s\n\n", utils3.IndentLines(spec.Description, "  ", false))
	}
	if spec.AdditionalResources == nil {
		out.Outf(a.cmd.Context, "no configuration templates found for %s in %s\n", ires.Meta().GetName(), nv)
		return nil
	}

	if len(spec.Scheme) > 0 && a.cmd.ParameterFile != "" {
		schemeFile := a.cmd.ParameterFile
		schemeFile = strings.TrimSuffix(schemeFile, ".yaml")
		schemeFile += ".jsonscheme"
		err = vfs.WriteFile(a.cmd.FileSystem(), schemeFile, spec.Scheme, 0o644)
		if err != nil {
			out.Errf(a.cmd.Context, "writing scheme file %s failed: %s\n", schemeFile, err)
		} else {
			out.Outf(a.cmd.Context, "%s: %d byte(s) written\n", schemeFile, len(spec.Scheme))
		}
	}

	list := errors.ErrList()
	list.Add(a.handle("configuration template", a.cmd.ParameterFile, cv, spec.AdditionalResources[toi.AdditionalResourceConfigFile], resolver))
	list.Add(a.handle("credentials template", a.cmd.CredentialsFile, cv, spec.AdditionalResources[toi.AdditionalResourceCredentialsFile], resolver))
	return list.Result()
}

func (a *action) handle(kind, path string, cv ocm.ComponentVersionAccess, spec *toi.AdditionalResource, resolver ocm.ComponentVersionResolver) error {
	var err error
	if spec != nil {
		if spec.ResourceReference != nil && len(spec.ResourceReference.Resource) != 0 {
			return a.download(kind, path, cv, spec.ResourceReference, resolver)
		} else {
			var content interface{}
			if len(spec.Content) > 0 {
				if err = json.Unmarshal(spec.Content, &content); err != nil {
					return errors.Wrapf(err, "cannot marshal %s content", kind)
				}
				l := 0
				out.Outf(a.cmd.Context, "writing %s...\n", kind)
				switch c := content.(type) {
				case string:
					l = len(c)
					err = vfs.WriteFile(a.cmd.FileSystem(), path, []byte(c), 0o600)
				case []byte:
					l = len(c)
					err = vfs.WriteFile(a.cmd.FileSystem(), path, c, 0o600)
				default:
					var data []byte
					data, err = runtime.DefaultYAMLEncoding.Marshal(spec.Content)
					if err != nil {
						data = spec.Content
					}
					l = len(spec.Content)
					err = vfs.WriteFile(a.cmd.FileSystem(), path, data, 0o600)
				}
				if err != nil {
					return errors.Wrapf(err, "cannot write %s to %s", kind, path)
				}
				out.Outf(a.cmd.Context, "%s: %d byte(s) written\n", path, l)
				return nil
			}
			return nil
		}
	}
	out.Outf(a.cmd.Context, "no %s configured\n", kind)
	return nil
}

func (a *action) download(kind, path string, cv ocm.ComponentVersionAccess, spec *metav1.ResourceReference, resolver ocm.ComponentVersionResolver) error {
	res, _, err := resourcerefs.MatchResourceReference(cv, toi.TypeYAML, *spec, resolver)
	if err != nil {
		return errors.Wrapf(err, "%s resource", kind)
	}
	out.Outf(a.cmd.Context, "downloading %s...\n", kind)
	ok, _, err := download.For(a.cmd.Context).DownloadAsBlob(common.NewPrinter(a.cmd.StdOut()), res, path, a.cmd.FileSystem())
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no downloader configured for type %q", res.Meta().GetType())
	}
	return nil
}
