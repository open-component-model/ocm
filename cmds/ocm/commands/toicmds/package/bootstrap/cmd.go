package bootstrap

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/tools/toi"
	defaultd "ocm.software/ocm/api/ocm/tools/toi/drivers/default"
	"ocm.software/ocm/api/ocm/tools/toi/drivers/docker"
	"ocm.software/ocm/api/ocm/tools/toi/drivers/filesystem"
	"ocm.software/ocm/api/ocm/tools/toi/install"
	utils2 "ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/listformat"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/out"
	"ocm.software/ocm/api/utils/runtime"
	ocmcommon "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/handlers/comphdlr"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/lookupoption"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/repooption"
	"ocm.software/ocm/cmds/ocm/commands/toicmds/names"
	"ocm.software/ocm/cmds/ocm/commands/verbs"
	"ocm.software/ocm/cmds/ocm/common/output"
	"ocm.software/ocm/cmds/ocm/common/utils"
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
	Credentials     blobaccess.DataSource
	Parameters      blobaccess.DataSource
	Config          map[string]string
	EnvDir          string
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
of software. See also the topic <CMD>ocm toi-bootstrapping</CMD>.

This resource finally describes an executor image, which will be executed in a
container with the installation source and (instance specific) user settings.
The container is just executed, the framework make no assumption about the
meaning/outcome of the execution. Therefore, any kind of actions can be described and
issued this way, not only installation handling.

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

One of <code>consumerId</code>, <code>reference</code> or <code>credentials</code> must be configured.

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
` + listformat.FormatListElements("", listformat.StringElementList(utils2.StringMapKeys(docker.Options))) + `

Using the option <code>--create-env  &lt;toi root folder></code> it is possible to
create a local execution environment for an executor according to the executor
image contract (see <CMD>ocm toi-bootstrapping</CMD>). If the executor executable is
built based on the toi executor support package, the executor can then be called
locally with

<center>
    <pre>&lt;executor> --bootstraproot &lt;given toi root folder></pre>
</center>
`,
		Example: `
$ ocm toi bootstrap package ghcr.io/mandelsoft/ocm//ocmdemoinstaller:0.0.1-dev
`,
		Annotations: map[string]string{"ExampleCodeStyle": "bash"},
	}
	return cmd
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	o.BaseCommand.AddFlags(fs)
	fs.StringToStringVarP(&o.Config, "config", "", nil, "driver config")
	fs.StringVarP(&o.CredentialsFile, "credentials", "c", "", "credentials file")
	fs.StringVarP(&o.ParameterFile, "parameters", "p", "", "parameter file")
	fs.StringVarP(&o.OutputFile, "outputs", "o", "", "output file/directory")
	fs.StringVarP(&o.EnvDir, "create-env", "C", "", "create local filesystem contract to call executor command locally")
}

func (o *Command) Complete(args []string) error {
	o.Action = args[0]
	o.Ref = args[1]
	id, err := ocmcommon.MapArgsToIdentityPattern(args[2:]...)
	if err != nil {
		return errors.Wrapf(err, "bootstrap resource identity pattern")
	}
	if len(o.CredentialsFile) == 0 {
		ok, err := vfs.FileExists(o.FileSystem(), DEFAULT_CREDENTIALS_FILE)
		if err != nil {
			return err
		}

		if ok {
			o.CredentialsFile = DEFAULT_CREDENTIALS_FILE
		}
	}
	o.Id = id
	if len(o.CredentialsFile) > 0 {
		data, err := utils2.ReadFile(o.CredentialsFile, o.Context.FileSystem())
		if err != nil {
			return errors.Wrapf(err, "failed reading credentials file %q", o.CredentialsFile)
		}
		o.Credentials = blobaccess.DataAccessForData(data, o.CredentialsFile)
	}
	if len(o.ParameterFile) == 0 {
		if ok, _ := vfs.FileExists(o.FileSystem(), DEFAULT_PARAMETER_FILE); ok {
			o.ParameterFile = DEFAULT_PARAMETER_FILE
		}
	}
	if len(o.ParameterFile) > 0 {
		data, err := utils2.ReadFile(o.ParameterFile, o.Context.FileSystem())
		if err != nil {
			return errors.Wrapf(err, "failed reading parameter file %q", o.ParameterFile)
		}
		o.Parameters = blobaccess.DataAccessForData(data, o.ParameterFile)
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

	if a.cmd.EnvDir != "" {
		driver = filesystem.New(a.cmd.FileSystem())
		if a.cmd.Config == nil {
			a.cmd.Config = map[string]string{}
		}
		if a.cmd.Config[filesystem.OptionTargetPath] == "" {
			a.cmd.Config[filesystem.OptionTargetPath] = a.cmd.EnvDir
		}
	}

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
