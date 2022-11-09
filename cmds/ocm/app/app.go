// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=mod ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"strings"

	_ "github.com/open-component-model/ocm/pkg/contexts/clictx/config"
	_ "github.com/open-component-model/ocm/pkg/contexts/ocm/attrs"

	"github.com/mandelsoft/logging"
	"github.com/mandelsoft/logging/config"
	"github.com/mandelsoft/logging/logrusr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/cachecmds"
	creds "github.com/open-component-model/ocm/cmds/ocm/commands/misccmds/credentials"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds"
	common2 "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/plugins"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/toicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/bootstrap"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/clean"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/show"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/sign"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/commands/verbs/verify"
	cmdutils "github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/cmds/ocm/topics/common/attributes"
	topicconfig "github.com/open-component-model/ocm/cmds/ocm/topics/common/config"
	topicocirefs "github.com/open-component-model/ocm/cmds/ocm/topics/oci/refs"
	topicocmrefs "github.com/open-component-model/ocm/cmds/ocm/topics/ocm/refs"
	topicbootstrap "github.com/open-component-model/ocm/cmds/ocm/topics/toi/bootstrapping"
	"github.com/open-component-model/ocm/pkg/cobrautils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	datacfg "github.com/open-component-model/ocm/pkg/contexts/datacontext/config/attrs"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/errors"
	ocmlog "github.com/open-component-model/ocm/pkg/logging"
	"github.com/open-component-model/ocm/pkg/out"
	"github.com/open-component-model/ocm/pkg/version"
)

type CLIOptions struct {
	Config      string
	Credentials []string
	Context     clictx.Context
	Settings    []string
	Verbose     bool
	LogLevel    string
	LogFile     string
	LogConfig   string

	logFile vfs.File
}

var desc = `
The Open Component Model command line client support the work with OCM
artefacts, like Component Archives, Common Transport Archive,  
Component Repositories, and component versions.

Additionally it provides some limited support for the docker daemon, OCI artefacts and
registries.

It can be used in two ways:
- *verb/operation first*: here the sub commands follow the pattern *&lt;verb> &lt;object kind> &lt;arguments>*
- *area/kind first*: here the area and/or object kind is given first followed by the operation according to the pattern
  *[&lt;area>] &lt;object kind> &lt;verb/operation> &lt;arguments>*

The command accepts some top level options, they can only be given before the sub commands.

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
    <pre>--cred :type=ociRegistry --cred hostname=ghcr.io --cred usename=mandelsoft --cred password=xyz</pre>
</center>

With the option <code>-X</code> it is possible to pass global settings of the 
form 

<center>
    <pre>-X &lt;attribute>=&lt;value></pre>
</center>

The value can be a simple type or a json string for complex values. The following
attributes are supported:
` + attributes.Attributes()

func NewCliCommandForArgs(ctx clictx.Context, args []string, mod ...func(clictx.Context, *cobra.Command)) (*cobra.Command, error) {
	opts, args, err := Prepare(ctx, args)
	if err != nil {
		return nil, err
	}
	cmd := newCliCommand(opts, mod...)
	cmd.SetArgs(args)
	return cmd, nil
}

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

	cmd.AddCommand(get.NewCommand(opts.Context))
	cmd.AddCommand(create.NewCommand(opts.Context))
	cmd.AddCommand(add.NewCommand(opts.Context))
	cmd.AddCommand(sign.NewCommand(opts.Context))
	cmd.AddCommand(verify.NewCommand(opts.Context))
	cmd.AddCommand(show.NewCommand(opts.Context))
	cmd.AddCommand(transfer.NewCommand(opts.Context))
	cmd.AddCommand(describe.NewCommand(opts.Context))
	cmd.AddCommand(download.NewCommand(opts.Context))
	cmd.AddCommand(bootstrap.NewCommand(opts.Context))
	cmd.AddCommand(clean.NewCommand(opts.Context))

	cmd.AddCommand(cmdutils.HideCommand(componentarchive.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(resources.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(references.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(sources.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(components.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(plugins.NewCommand(opts.Context)))

	cmd.AddCommand(cmdutils.HideCommand(cachecmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(ocicmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(ocmcmds.NewCommand(opts.Context)))
	cmd.AddCommand(cmdutils.HideCommand(toicmds.NewCommand(opts.Context)))

	cmd.AddCommand(cmdutils.HideCommand(creds.NewCommand(opts.Context)))

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
	cmd.AddCommand(topicocirefs.New(ctx))
	cmd.AddCommand(topicocmrefs.New(ctx))
	cmd.AddCommand(attributes.New(ctx))
	cmd.AddCommand(topicbootstrap.New(ctx, "toi-bootstrapping"))

	help.AddCommand(topicconfig.New(ctx))
	help.AddCommand(topicocirefs.New(ctx))
	help.AddCommand(topicocmrefs.New(ctx))
	help.AddCommand(topicbootstrap.New(ctx, "toi-bootstrapping"))

	for _, m := range mod {
		if m != nil {
			m(ctx, cmd)
		}
	}
	return cmd
}

func (o *CLIOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Config, "config", "", "", "configuration file")
	fs.StringArrayVarP(&o.Credentials, "cred", "C", nil, "credential setting")
	fs.StringArrayVarP(&o.Settings, "attribute", "X", nil, "attribute setting")
	fs.BoolVarP(&o.Verbose, "verbose", "v", false, "enable verbose logging")
	fs.StringVarP(&o.LogLevel, "loglevel", "l", "", "set log level")
	fs.StringVarP(&o.LogFile, "logfile", "L", "", "set log file")
	fs.StringVarP(&o.LogConfig, "logconfig", "", "", "log config")
}

func (o *CLIOptions) Close() error {
	if o.logFile == nil {
		return nil
	}
	return o.logFile.Close()
}

func (o *CLIOptions) Complete() error {
	var err error
	if o.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if o.LogLevel != "" {
		l, err := logging.ParseLevel(o.LogLevel)
		if err != nil {
			return errors.Wrapf(err, "invalid log level %q", o.LogLevel)
		}
		ocmlog.Context().SetDefaultLevel(l)
	} else {
		ocmlog.Context().SetDefaultLevel(logging.ErrorLevel)
	}

	if o.LogFile != "" {
		o.logFile, err = o.Context.FileSystem().OpenFile(o.LogFile, vfs.O_CREATE|vfs.O_WRONLY, 0o600)
		if err != nil {
			return errors.Wrapf(err, "cannot open log file %q", o.LogFile)
		}
		log := logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
		log.SetOutput(o.logFile)
		ocmlog.Context().SetBaseLogger(logrusr.New(log))
	}

	if o.LogConfig != "" {
		cfg, err := vfs.ReadFile(o.Context.FileSystem(), o.LogConfig)
		if err != nil {
			return errors.Wrapf(err, "cannot read logging config %q", o.LogFile)
		}
		if err = config.ConfigureWithData(ocmlog.Context(), cfg); err != nil {
			return errors.Wrapf(err, "cinvalid logging config: %q", o.LogFile)
		}
	}

	_, err = utils.Configure(o.Context.OCMContext(), o.Config, vfsattr.Get(o.Context))
	if err != nil {
		return err
	}

	id := credentials.ConsumerIdentity{}
	attrs := common.Properties{}
	for _, s := range o.Credentials {
		i := strings.Index(s, "=")
		if i < 0 {
			return errors.ErrInvalid("credential setting", s)
		}
		name := s[:i]
		value := s[i+1:]
		if strings.HasPrefix(name, ":") {
			if len(attrs) != 0 {
				o.Context.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
				id = credentials.ConsumerIdentity{}
				attrs = common.Properties{}
			}
			name = name[1:]
			id[name] = value
		} else {
			attrs[name] = value
		}
		if len(name) == 0 {
			return errors.ErrInvalid("credential setting", s)
		}
	}
	if len(attrs) != 0 {
		o.Context.CredentialsContext().SetCredentialsForConsumer(id, credentials.NewCredentials(attrs))
	} else if len(id) != 0 {
		return errors.Newf("empty credential attribute set for %s", id.String())
	}

	set, err := common2.ParseLabels(o.Settings, "attribute setting")
	if err == nil && len(set) > 0 {
		ctx := o.Context.ConfigContext()
		spec := datacfg.New()
		for _, s := range set {
			attr := s.Name
			eff := datacontext.DefaultAttributeScheme.Shortcuts()[attr]
			if eff != "" {
				attr = eff
			}
			err = spec.AddRawAttribute(attr, s.Value)
			if err != nil {
				return errors.Wrapf(err, "attribute %s", s.Name)
			}
		}
		err = ctx.ApplyConfig(spec, "cli")
	}
	return plugincacheattr.Get(o.Context.OCMContext()).RegisterExtensions()
}

func NewVersionCommand(ctx clictx.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			out.Outf(ctx, "%#v\n", v)
		},
	}
}
