// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=mod ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/open-component-model/ocm/cmds/ocm/commands/add"
	"github.com/open-component-model/ocm/cmds/ocm/commands/create"
	"github.com/open-component-model/ocm/cmds/ocm/commands/describe"
	"github.com/open-component-model/ocm/cmds/ocm/commands/download"
	"github.com/open-component-model/ocm/cmds/ocm/commands/get"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocicmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/componentarchive"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/components"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/references"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/resources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/sources"
	"github.com/open-component-model/ocm/cmds/ocm/commands/transfer"
	"github.com/open-component-model/ocm/pkg/contexts/config"

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
	Config  string
	Context clictx.Context
}

func NewCliCommand(ctx clictx.Context) *cobra.Command {
	if ctx == nil {
		ctx = clictx.DefaultContext()
	}
	opts := &CLIOptions{
		Context: ctx,
	}
	cmd := &cobra.Command{
		Use:              "ocm",
		Short:            "ocm",
		TraverseChildren: true,
		Version:          version.Get().String(),
		SilenceUsage:     true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Complete()
		},
	}

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(get.NewCommand(opts.Context))
	cmd.AddCommand(create.NewCommand(opts.Context))
	cmd.AddCommand(add.NewCommand(opts.Context))
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
