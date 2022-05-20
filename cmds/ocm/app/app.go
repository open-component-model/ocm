// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=mod ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	dockercli "github.com/docker/cli/cli/config"
	"github.com/open-component-model/ocm/cmds/ocm/commands/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds"
	common2 "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/show"
	"github.com/open-component-model/ocm/cmds/ocm/commands/transfer"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/cobrautils"
	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	credcfg "github.com/open-component-model/ocm/pkg/contexts/credentials/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials/repositories/dockerconfig"
	config2 "github.com/open-component-model/ocm/pkg/contexts/datacontext/config"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/pkg/errors"
	"github.com/open-component-model/ocm/pkg/version"
	"github.com/spf13/pflag"

	_ "github.com/open-component-model/ocm/cmds/ocm/clictx/config"
)

type CLI struct {
	clictx.Context
}

func NewCLI(ctx clictx.Context) *CLI {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	return &CLI{ctx}
}

func (c *CLI) Execute(args ...string) error {
	cmd := NewCliCommand(c)
	cmd.SetArgs(args)
	return cmd.Execute()
}

type CLIOptions struct {
	Config      string
	Credentials []string
	Context     clictx.Context
	Settings    []string
}

var desc = `
The Open Component Model command line client support the work with OCM
artefacts, like Component Archives, Common Transport Archive,  
Component Repositories, and component versions.

Additionally it provides some limited support for the docker daemon, OCI artefacts and
registries.

With the open <code>--cred</code> it is possible to specify arbitrary credentials
for various environments on the command line. Nevertheless it is always preferrable
to use the cli config file.
Every credential setting is related to a dedicated consumer and provides a set of
credential attributes. All this can be specified by a sequence of <code>--cred</code>
options. 

Every option value has the format
<center>
<code>--cred [:]&lt;attr>=&lt;value></code>
</center>

Consumer identity attribues are prefixed with the colon (:). A credential settings
always start with a sequence of at least one identity attributes, followed by a
sequence of credential attributes.
If a credential attribute is followed by an identity attribute a new credential setting
is started.

The first credential setting may omit identity attributes. In this case it is used as
default credential, always used if no dedicated match is found.

For example:
<center>
<code>--cred :type=ociRegistry --cred hostname:ghcr.io --cred usename=mandelsoft --cred password=xyz </code>
</center>
`

func NewCliCommand(ctx clictx.Context) *cobra.Command {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	opts := &CLIOptions{
		Context: ctx,
	}
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
	}
	cobrautils.TweakCommand(cmd, ctx)

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(get.NewCommand(opts.Context))
	cmd.AddCommand(create.NewCommand(opts.Context))
	cmd.AddCommand(add.NewCommand(opts.Context))
	cmd.AddCommand(show.NewCommand(opts.Context))
	cmd.AddCommand(transfer.NewCommand(opts.Context))
	cmd.AddCommand(describe.NewCommand(opts.Context))
	cmd.AddCommand(download.NewCommand(opts.Context))

	cmd.AddCommand(componentarchive.NewCommand(opts.Context))
	cmd.AddCommand(resources.NewCommand(opts.Context))
	cmd.AddCommand(references.NewCommand(opts.Context))
	cmd.AddCommand(sources.NewCommand(opts.Context))
	cmd.AddCommand(components.NewCommand(opts.Context))

	cmd.AddCommand(ocicmds.NewCommand(opts.Context))
	cmd.AddCommand(ocmcmds.NewCommand(opts.Context))

	opts.AddFlags(cmd.Flags())
	return cmd
}

func (o *CLIOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Config, "config", "", "", "configuration file")
	fs.StringArrayVarP(&o.Credentials, "cred", "C", nil, "credential setting")
	fs.StringArrayVarP(&o.Settings, "", "X", nil, "attribute setting")
}

func (o *CLIOptions) Complete() error {
	h := os.Getenv("HOME")
	if o.Config == "" {
		if h != "" {
			cfg := h + "/.ocmconfig"
			if ok, err := vfs.FileExists(osfs.New(), cfg); ok && err == nil {
				o.Config = cfg
			}
		}
	}
	if o.Config != "" {
		//fmt.Printf("********** config file is %s\n", o.Config)
		if strings.HasPrefix(o.Config, "~"+string(os.PathSeparator)) {
			if len(h) == 0 {
				return fmt.Errorf("no home directory found for resolving path of config file %q", o.Config)
			}
			o.Config = h + o.Config[1:]
		}
		data, err := ioutil.ReadFile(o.Config)
		if err != nil {
			return errors.Wrapf(err, "cannot read config file %q", o.Config)
		}

		cfg, err := config.DefaultContext().GetConfigForData(data, nil)
		if err != nil {
			return errors.Wrapf(err, "invalid config file %q", o.Config)
		}
		o.Context = clictx.DefaultContext()
		err = config.DefaultContext().ApplyConfig(cfg, o.Config)
		if err != nil {
			return errors.Wrapf(err, "cannot apply config %q", o.Config)
		}
	} else {
		// use docker config as default config for ocm cli
		d := filepath.Join(dockercli.Dir(), dockercli.ConfigFileName)
		if ok, err := vfs.FileExists(osfs.New(), d); ok && err == nil {
			cfg := credcfg.NewConfigSpec()
			cfg.AddRepository(dockerconfig.NewRepositorySpec(d, true))
			err = config.DefaultContext().ApplyConfig(cfg, d)
			if err != nil {
				return errors.Wrapf(err, "cannot apply docker config %q", d)
			}
		}
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
		id = credentials.ConsumerIdentity{}
		attrs = common.Properties{}
	} else {
		if len(id) != 0 {
			return errors.Newf("empty credential attribute set for %s", id.String())
		}
	}

	set, err := common2.ParseLabels(o.Settings, "attribute setting")
	if err != nil {
		return err
	}
	spec := config2.NewConfigSpec()
	for _, s := range set {
		err = spec.AddRawAttribute(s.Name, s.Value)
		if err != nil {
			return errors.Wrapf(err, "attribute %s", s.Name)
		}
	}
	return nil
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Printf("%#v", v)
		},
	}
}
