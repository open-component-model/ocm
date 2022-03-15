// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=vendor ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gardener/ocm/cmds/ocm/commands/create"
	"github.com/gardener/ocm/cmds/ocm/commands/get"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"

	"github.com/gardener/ocm/cmds/ocm/cmd"
	"github.com/gardener/ocm/pkg/config"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/version"
	"github.com/spf13/pflag"

	_ "github.com/gardener/ocm/cmds/ocm/cmd/config"
)

type CLIOptions struct {
	Config  string
	Context cmd.Context
}

func NewCliCommand(ctx context.Context) *cobra.Command {
	opts := &CLIOptions{
		Context: cmd.DefaultContext(),
	}
	cmd := &cobra.Command{
		Use:              "ocm",
		Short:            "ocm",
		TraverseChildren: true,
		Version:          version.Get().String(),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(get.NewCommand(opts.Context))
	cmd.AddCommand(create.NewCommand(opts.Context))

	opts.AddFlags(cmd.Flags())
	return cmd
}

func (o *CLIOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Config, "config", "", "", "configuration file")
}

func (o *CLIOptions) Complete() error {
	h := os.Getenv("HOME")
	fmt.Printf("found config file %q\n", o.Config)

	if o.Config == "" {
		if h != "" {
			cfg := h + "/.ocmconfig"
			if ok, err := vfs.FileExists(osfs.New(), cfg); ok && err != nil {
				o.Config = cfg
			}
		}
	}
	if o.Config != "" {
		fmt.Printf("********** config file is %s\n", o.Config)
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
		o.Context = cmd.DefaultContext()
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
