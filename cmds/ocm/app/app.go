//go:generate go run -mod=mod ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"fmt"
	"strings"
	"unicode"

	_ "github.com/open-component-model/ocm/pkg/contexts/clictx/config"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/attrs"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	config2 "github.com/open-component-model/ocm/cmds/ocm/clippi/config"
	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/common/options/keyoption"
	"github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/action"
	creds "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/routingslips"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/plugin"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/check"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/clean"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/controller"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/execute"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/hash"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/install"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/list"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/show"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/verify"
	cmdutils "github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/cmds/ocm/topics/common/attributes"
	topicconfig "github.com/open-component-model/ocm/cmds/ocm/topics/common/config"
	topiccredentials "github.com/open-component-model/ocm/cmds/ocm/topics/common/credentials"
	topiclogging "github.com/open-component-model/ocm/cmds/ocm/topics/common/logging"
	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
	topicocmaccessmethods "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/accessmethods"
	topicocmdownloaders "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/downloadhandlers"
	topicocmlabels "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/labels"
	topicocmrefs "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/refs"
	topicocmuploaders "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/uploadhandlers"
	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/toi/bootstrapping"
	"github.com/open-component-model/ocm/pkg/cobrautils"
	"github.com/open-component-model/ocm/pkg/cobrautils/logopts"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/clicfgattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/defaultconfigregistry"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/version"
)

type CLIOptions struct {
	config2.Config
	Completed bool
	Version   bool

	*config2.EvaluatedOptions
	Context clictx.Context
}

var desc = `
The Open Component Model command line client supports the work with OCM
artifacts, like Component Archives, Common Transport Archive,
Component Repositories, and Component Versions.

Additionally it provides some limited support for the docker daemon, OCI artifacts and
registries.

It can be used in two ways:
- *verb/operation first*: here the sub commands follow the pattern *&lt;verb> &lt;object kind> &lt;arguments>*
- *area/kind first*: here the area and/or object kind is given first followed by the operation according to the pattern
  *[&lt;area>] &lt;object kind> &lt;verb/operation> &lt;arguments>*

The command accepts some top level options, they can only be given before the sub commands.

A configuration according to <CMD>ocm configfile</CMD> is read from a <code>.ocmconfig</code> file
located in the <code>HOME</code> directory. With the option <code>--config</code> other
file locations can be specified. If nothing is specified and no file is found at the default
location a default configuration is composed according to known type specific
configuration files.

The following configuration sources are used:
` + defaultconfigregistry.Description() + `

With the option <code>--cred</code> it is possible to specify arbitrary credentials
for various environments on the command line. Nevertheless it is always preferrable
to use the cli config file.
Every credential setting is related to a dedicated consumer and provides a set of
credential attributes. All this can be specified by a sequence of <code>--cred</code>
options.

Every option value has the format

<center>
    <pre>--cred [:]&lt;attr>=&lt;value></pre>
</center>

Consumer identity attributes are prefixed with the colon (:). A credential settings
always start with a sequence of at least one identity attributes, followed by a
sequence of credential attributes.
If a credential attribute is followed by an identity attribute a new credential setting
is started.

The first credential setting may omit identity attributes. In this case it is used as
default credential, always used if no dedicated match is found.

For example:

<center>
    <pre>--cred :type=OCIRegistry --cred :hostname=ghcr.io --cred username=mandelsoft --cred password=xyz</pre>
</center>

With the option <code>-X</code> it is possible to pass global settings of the
form

<center>
    <pre>-X &lt;attribute>=&lt;value></pre>
</center>
` + logopts.Description + `
The value can be a simple type or a JSON/YAML string for complex values
(see <CMD>ocm attributes</CMD>. The following attributes are supported:
` + attributes.Attributes() + `

For several options (like <code>-X</code>) it is possible to pass complex values
using JSON or YAML syntax. To pass those arguments the escaping of the used shell
must be used to pass quotes, commas, curly brackets or newlines. for the *bash*
the easiest way to achieve this is to put the complete value into single quotes.

<center>
<code>-X 'mapocirepo={"mode": "shortHash"}'</code>.
</center>

Alternatively, quotes and opening curly brackets can be escaped by using a
backslash (<code>&bsol;</code>).
Often a tagged value can also be substituted from a file with the syntax

<center>
<code>&lt;attr>=@&lt;filepath></code>
</center>
` + keyoption.Usage()

// NewCliCommandForArgs is the regular way to instantiate a new CLI command.
// It observes settings provides by options for the main command.
// This especially means, that help texts are configured according
// to the config settings provided by options.
func NewCliCommandForArgs(ctx clictx.Context, args []string, mod ...func(clictx.Context, *cobra.Command)) (*cobra.Command, error) {
	for _, m := range mod {
		m(ctx, nil)
	}
	opts, args, err := Prepare(ctx, args)
	if err != nil {
		return nil, err
	}
	clicfgattr.Set(ctx.OCMContext(), opts.ConfigForward)
	cmd := newCliCommand(opts, mod...)
	cmd.SetArgs(args)
	return cmd, nil
}

// NewCliCommand creates a new command WITHOUT observing configuration options.
// The result is a command configured by pure defaults. This is especially true
// for plugin settings.
func NewCliCommand(ctx clictx.Context, mod ...func(clictx.Context, *cobra.Command)) *cobra.Command {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	opts := &CLIOptions{
		Context: ctx,
	}
	return newCliCommand(opts, append(mod, func(_ clictx.Context, cmd *cobra.Command) { cmd.PersistentPreRun = nil })...)
}

func newCliCommand(opts *CLIOptions, mod ...func(clictx.Context, *cobra.Command)) *cobra.Command {
	if opts == nil {
		opts = &CLIOptions{}
	}
	if opts.Context == nil {
		opts.Context = clictx.DefaultContext()
	}

	ctx := opts.Context
	cmd := &cobra.Command{
		Use:                   "ocm",
		Short:                 "Open Component Model command line client",
		Long:                  desc,
		Version:               version.Get().String(),
		TraverseChildren:      true,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete()
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Close()
		},
	}
	cobrautils.TweakCommand(cmd, ctx)

	cmd.AddCommand(NewVersionCommand(opts.Context))

	cmd.AddCommand(check.NewCommand(opts.Context))
	cmd.AddCommand(get.NewCommand(opts.Context))
	cmd.AddCommand(list.NewCommand(opts.Context))
	cmd.AddCommand(create.NewCommand(opts.Context))
	cmd.AddCommand(add.NewCommand(opts.Context))
	cmd.AddCommand(sign.NewCommand(opts.Context))
	cmd.AddCommand(hash.NewCommand(opts.Context))
	cmd.AddCommand(verify.NewCommand(opts.Context))
	cmd.AddCommand(show.NewCommand(opts.Context))
	cmd.AddCommand(transfer.NewCommand(opts.Context))
	cmd.AddCommand(describe.NewCommand(opts.Context))
	cmd.AddCommand(download.NewCommand(opts.Context))
	cmd.AddCommand(bootstrap.NewCommand(opts.Context))
	cmd.AddCommand(clean.NewCommand(opts.Context))
	cmd.AddCommand(install.NewCommand(opts.Context))
	cmd.AddCommand(execute.NewCommand(opts.Context))
	cmd.AddCommand(controller.NewCommand(opts.Context))

	cmd.AddCommand(cmdutils.HideCommand(componentarchive.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(resources.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(references.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(sources.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(components.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(plugins.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(action.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(routingslips.NewCommand(opts.Context)))

	cmd.AddCommand(cmdutils.OverviewCommand(cachecmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.OverviewCommand(ocicmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.OverviewCommand(ocmcmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.OverviewCommand(toicmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.OverviewCommand(creds.NewCommand(opts.Context)))

	opts.AddFlags(cmd.Flags())
	cmd.InitDefaultHelpCmd()
	var help *cobra.Command
	for _, c := range cmd.Commands() {
		if c.Name() == "help" {
			help = c
			break
		}
	}
	// help.Use="help <topic>"
	help.DisableFlagsInUseLine = true
	cmd.AddCommand(topicconfig.New(ctx))
	cmd.AddCommand(topiccredentials.New(ctx))
	cmd.AddCommand(topiclogging.New(ctx))
	cmd.AddCommand(topicocirefs.New(ctx))
	cmd.AddCommand(topicocmrefs.New(ctx))
	cmd.AddCommand(topicocmaccessmethods.New(ctx))
	cmd.AddCommand(topicocmuploaders.New(ctx))
	cmd.AddCommand(topicocmdownloaders.New(ctx))
	cmd.AddCommand(topicocmlabels.New(ctx))
	cmd.AddCommand(attributes.New(ctx))
	cmd.AddCommand(topicbootstrap.New(ctx, "toi-bootstrapping"))

	help.AddCommand(topicconfig.New(ctx))
	help.AddCommand(topiccredentials.New(ctx))
	help.AddCommand(topiclogging.New(ctx))
	help.AddCommand(topicocirefs.New(ctx))
	help.AddCommand(topicocmrefs.New(ctx))
	help.AddCommand(topicocmaccessmethods.New(ctx))
	help.AddCommand(topicocmuploaders.New(ctx))
	help.AddCommand(topicocmdownloaders.New(ctx))
	help.AddCommand(topicocmlabels.New(ctx))
	help.AddCommand(attributes.New(ctx))
	help.AddCommand(topicbootstrap.New(ctx, "toi-bootstrapping"))

	// register CLI extension commands
	pi := plugincacheattr.Get(ctx)

	for _, n := range pi.PluginNames() {
		p := pi.Get(n)
		if !p.IsValid() {
			continue
		}
		for _, c := range p.GetDescriptor().Commands {
			if c.Verb != "" {
				v := cobrautils.Find(cmd, c.Verb)
				if v == nil {
					out.Errf(opts.Context, "unknown verb %q for cli command %q of plugin %q", c.Verb, c.Name, p.Name())
				} else {
					s := cobrautils.Find(v, c.Name)
					if s != nil {
						out.Errf(opts.Context, "duplicate cli command %q of plugin %q for verb %q", c.Name, p.Name(), c.Verb)
					} else {
						v.AddCommand(plugin.NewCommand(ctx, p, c.Name))
					}
				}

				if c.Realm != "" {
					r := cobrautils.Find(cmd, c.Realm)
					if r == nil {
						out.Errf(opts.Context, "unknown realm %q for cli command %q of plugin %q", c.Realm, c.Name, p.Name())
					} else {
						v := cobrautils.Find(r, c.Name)
						if v == nil {
							out.Errf(opts.Context, "unknown object %q for cli command %q of plugin %q", c.Realm, c.Name, p.Name())
						} else {
							s := cobrautils.Find(v, c.Verb)
							if s != nil {
								out.Errf(opts.Context, "duplicate cli command %q of plugin %q for realm %q verb %q", c.Name, p.Name(), c.Realm, c.Verb)
							} else {
								v.AddCommand(plugin.NewCommand(ctx, p, c.Verb))
							}
						}
					}
				}
			} else {
				s := cobrautils.Find(cmd, c.Name)
				if s != nil {
					out.Errf(opts.Context, "duplicate top-level cli command %q of plugin %q", c.Name, p.Name())
				} else {
					cmd.AddCommand(plugin.NewCommand(ctx, p, c.Name))
				}
			}
		}
	}

	for _, m := range mod {
		if m != nil {
			m(ctx, cmd)
		}
	}
	return cmd
}

func (o *CLIOptions) AddFlags(fs *pflag.FlagSet) {
	o.Config.AddFlags(fs)

	fs.BoolVarP(&o.Version, "version", "", false, "show version") // otherwise it is implicitly added by cobra
}

func (o *CLIOptions) Close() error {
	return o.LogOpts.Close()
}

func (o *CLIOptions) Complete() error {
	var err error

	if o.Completed {
		return nil
	}
	o.Completed = true

	if o.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	old := o.Context.ConfigContext().SkipUnknownConfig(true)
	defer o.Context.ConfigContext().SkipUnknownConfig(old)

	o.EvaluatedOptions, err = o.Config.Evaluate(o.Context.OCMContext())
	if err != nil {
		return err
	}
	err = registration.RegisterExtensions(o.Context)
	if err != nil {
		return err
	}
	return o.Context.ConfigContext().Validate()
}

func prepare(s string) string {
	idx := 0
	for {
		i := strings.Index(s[idx:], ":\"")
		if i < 0 {
			break
		}
		idx += i
		j := idx - 1
		for unicode.IsNumber(rune(s[j])) || unicode.IsLetter(rune(s[j])) {
			j--
		}
		if j != idx-1 {
			s = s[:j+1] + "\"" + s[j+1:idx] + "\"" + s[idx:]
			idx += 2
		}
		idx += 1
	}
	return s
}

func NewVersionCommand(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(_ *cobra.Command, _ []string) {
			v := version.Get()
			s := fmt.Sprintf("%#v", v)
			out.Outf(ctx, "%s\n", prepare(s[strings.Index(s, "{"):])) //nolint:gocritic // yes
		},
	}
}
