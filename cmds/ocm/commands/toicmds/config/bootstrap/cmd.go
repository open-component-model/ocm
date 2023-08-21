// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/package/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/toi/bootstrapping"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/ociuploadattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	utils2 "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi"
	"github.com/open-component-model/ocm/pkg/toi/install"
	utils3 "github.com/open-component-model/ocm/pkg/utils"
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
of software. See also the topic <CMD>ocm toi toi-bootstrapping</CMD>.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no credentials file name is provided (option -c) the file
<code>` + DEFAULT_CREDENTIALS_FILE + `</code> is used. If no parameter file name is
provided (option -p) the file <code>` + DEFAULT_PARAMETER_FILE + `</code> is used.

For more details about those files see <CMD> ocm bootstrap package</CMD>.
`,
		Example: `
$ ocm toi bootstrap config ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
`,
	}
	cmd.AddCommand(topicbootstrap.New(o.Context, "toi-bootstrapping"))
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

	ires, eff, err := utils2.MatchResourceReference(cv, toi.TypeTOIPackage, rid, resolver)
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

	err = nil
	if len(spec.Scheme) > 0 && a.cmd.ParameterFile != "" {
		schemeFile := a.cmd.ParameterFile
		if strings.HasSuffix(schemeFile, ".yaml") {
			schemeFile = schemeFile[:len(schemeFile)-5]
		}
		schemeFile = schemeFile + ".jsonscheme"
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
					data, err := runtime.DefaultYAMLEncoding.Marshal(spec.Content)
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
	res, _, err := utils2.MatchResourceReference(cv, toi.TypeYAML, *spec, resolver)
	if err != nil {
		return errors.Wrapf(err, "%s resource", kind)
	}
	out.Outf(a.cmd.Context, "downloading %s...\n", kind)
	ok, path, err := download.For(a.cmd.Context).DownloadAsBlob(common.NewPrinter(a.cmd.StdOut()), res, path, a.cmd.FileSystem())
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no downloader configured for type %q", res.Meta().GetType())
	}
	return nil
}
