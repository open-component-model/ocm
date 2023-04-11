// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"fmt"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ocmcommon "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/output"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/toi/bootstrapping"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/runtime"
	"github.com/open-component-model/ocm/pkg/toi"
	defaultd "github.com/open-component-model/ocm/pkg/toi/drivers/default"
	"github.com/open-component-model/ocm/pkg/toi/drivers/docker"
	"github.com/open-component-model/ocm/pkg/toi/install"
	utils2 "github.com/open-component-model/ocm/pkg/utils"
)

const (
	DEFAULT_CREDENTIALS_FILE = "TOICredentials"
	DEFAULT_PARAMETER_FILE   = "TOIParameters"
)

var (
	Names = names.Package
	Verb  = verbs.Bootstrap
)

type Command struct {
	utils.BaseCommand
	Action string
	Ref    string
	Id     v1.Identity

	CredentialsFile string
	ParameterFile   string
	OutputFile      string
	Credentials     accessio.DataSource
	Parameters      accessio.DataSource
	Config          map[string]string
}

// NewCommand creates a new bootstrap component command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx, repooption.New(), lookupoption.New())}, utils.Names(Names, names...)...)
}

func (o *Command) ForName(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[<options>] <action> {<component-reference>} {<resource id field>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "bootstrap component version",
		Long: `
Use the simple TOI bootstrap mechanism to execute actions for a TOI package resource
based on the content of an OCM component version and some command input describing
the dedicated installation target.

The package resource must have the type <code>` + toi.TypeTOIPackage + `</code>.
This is a simple YAML file resource describing the bootstrapping of a dedicated kind
of software. See also the topic <CMD>ocm toi toi-bootstrapping</CMD>.

This resource finally describes an executor image, which will be executed in a
container with the installation source and (instance specific) user settings.
The container is just executed, the framework make no assumption about the
meaning/outcome of the execution. Therefore, any kind of actions can be described and
issued this way, not on installation handling.

The first matching resource of this type is selected. Optionally a set of
identity attribute can be specified used to refine the match. This can be the
resource name and/or other key/value pairs (<code>&lt;attr>=&lt;value></code>).

If no output file is provided, the yaml representation of the outputs are
printed to standard out. If the output file is a directory, for every output a
dedicated file is created, otherwise the yaml representation is stored to the
file.

If no credentials file name is provided (option -c) the file
<code>` + DEFAULT_CREDENTIALS_FILE + `</code> is used, if present. If no parameter file name is
provided (option -p) the file <code>` + DEFAULT_PARAMETER_FILE + `</code> is used, if present.

Using the credentials file it is possible to configure credentials required by
the installation package or executor. Additionally arbitrary consumer ids
can be forwarded to executor, which might be required by accessing blobs
described by external access methods.

The credentials file uses the following yaml format:
- <code>credentials</code> *map[string]CredentialsSpec*

  The resolution of credentials requested by the package (by name).

- <code>forwardedConsumers</code> *[]ForwardSpec* (optional)

  An optional list of consumer specifications to be forwarded to the OCM
  configuration provided to the executor.

The *CredentialsSpec* uses the following format:

- <code>consumerId</code> *map[string]string*

  The consumer id used to look up the credentials.

- <code>consumerType</code> *string* (optional) (default: partial)

  The type of the matcher used to match the consumer id.

- <code>reference</code> *yaml*

  A generic credential specification as used in the ocm config file.

- <code>credentials</code> *map[string]string*

  Direct credential fields.

One of <code>consumerId</code>, <code>reference</code> or <code>credentials</code>
must be configured.

The *ForwardSpec* uses the following format:

- <code>consumerId</code> *map[string]string*

  The consumer id to be forwarded.

- <code>consumerType</code> *string* (optional) (default: partial)

  The type of the matcher used to match the consumer id.

If provided by the package it is possible to download template versions
for the parameter and credentials file using the command <CMD>ocm bootstrap configuration</CMD>.

Using the option <code>--config</code> it is possible to configure options
for the execution environment (so far only docker is supported).
The following options are possible:
` + utils.FormatListElements("", utils.StringElementList(utils2.StringMapKeys(docker.Options))),
		Example: `
$ ocm toi bootstrap package ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
`,
	}
	cmd.AddCommand(topicbootstrap.New(o.Context, "toi-bootstrapping"))
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringToStringVarP(&o.Config, "config", "", nil, "driver config")
	fs.StringVarP(&o.CredentialsFile, "credentials", "c", "", "credentials file")
	fs.StringVarP(&o.ParameterFile, "parameters", "p", "", "parameter file")
	fs.StringVarP(&o.OutputFile, "outputs", "o", "", "output file/directory")
}

func (o *Command) Complete(args []string) error {
	o.Action = args[0]
	o.Ref = args[1]
	id, err := ocmcommon.MapArgsToIdentityPattern(args[2:]...)
	if err != nil {
		return errors.Wrapf(err, "bootstrap resource identity pattern")
	}
	if len(o.CredentialsFile) == 0 {
		if ok, _ := vfs.FileExists(o.FileSystem(), DEFAULT_CREDENTIALS_FILE); ok {
			o.CredentialsFile = DEFAULT_CREDENTIALS_FILE
		}
	}
	o.Id = id
	if len(o.CredentialsFile) > 0 {
		data, err := vfs.ReadFile(o.Context.FileSystem(), o.CredentialsFile)
		if err != nil {
			return errors.Wrapf(err, "failed reading credentials file %q", o.CredentialsFile)
		}
		o.Credentials = accessio.DataAccessForBytes(data, o.CredentialsFile)
	}
	if len(o.ParameterFile) == 0 {
		if ok, _ := vfs.FileExists(o.FileSystem(), DEFAULT_PARAMETER_FILE); ok {
			o.ParameterFile = DEFAULT_PARAMETER_FILE
		}
	}
	if len(o.ParameterFile) > 0 {
		data, err := vfs.ReadFile(o.Context.FileSystem(), o.ParameterFile)
		if err != nil {
			return errors.Wrapf(err, "failed reading parameter file %q", o.ParameterFile)
		}
		o.Parameters = accessio.DataAccessForBytes(data, o.ParameterFile)
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
		out.Outf(a.cmd, "Warning: repository is no OCI registry, consider importing it or use upload repository with option ' -X ociuploadrepo=...'")
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
	driver := defaultd.New()

	if a.cmd.Config != nil {
		err := driver.SetConfig(a.cmd.Config)
		if err != nil {
			return err
		}
	}

	common.NewPrinter(a.cmd.StdOut())
	result, err := install.Execute(common.NewPrinter(a.cmd.StdOut()), driver, a.cmd.Action, a.cmd.Id, a.cmd.Credentials, a.cmd.Parameters, a.cmd.OCMContext(), a.data[0].ComponentVersion, lookupoption.From(a.cmd))
	if err != nil {
		return err
	}

	if a.cmd.OutputFile != "" {
		if ok, _ := vfs.IsDir(a.cmd.FileSystem(), a.cmd.OutputFile); ok {
			out.Outf(a.cmd, "writing outputs to directory %q...", a.cmd.OutputFile)
			for n, o := range result.Outputs {
				err := vfs.WriteFile(a.cmd.FileSystem(), vfs.Join(a.cmd.FileSystem(), a.cmd.OutputFile, n), o, 0o600)
				if err != nil {
					return errors.Wrapf(err, "cannot write output %q", n)
				}
			}
			return nil
		}
	}

	data := map[string]interface{}{}
	for n, o := range result.Outputs {
		var tmp interface{}
		err := runtime.DefaultYAMLEncoding.Unmarshal(o, &tmp)
		if err == nil {
			data[n] = tmp
		} else {
			data[n] = &Binary{o}
		}
	}

	outputs, err := runtime.DefaultYAMLEncoding.Marshal(map[string]interface{}{"outputs": data})
	if err != nil {
		return errors.Wrapf(err, "cannot marshal outputs")
	}
	if a.cmd.OutputFile != "" {
		vfs.WriteFile(a.cmd.FileSystem(), a.cmd.OutputFile, outputs, 0o600)
	} else {
		out.Outf(a.cmd, "Provided outputs:\n%s\n", outputs)
	}
	return nil
}
