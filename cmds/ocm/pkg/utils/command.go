// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package utils

import (
	"os"
	"strings"

	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/options"
	"github.com/gardener/ocm/cmds/ocm/pkg/output/out"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// OCMCommand is a command pattern, thta can be instantiated for a dediated
// sub command name.
type OCMCommand interface {
	clictx.Context
	ForName(name string) *cobra.Command
	AddFlags(fs *pflag.FlagSet)
	Complete(args []string) error
	Run() error
}

type BaseCommand struct {
	clictx.Context
	options.OptionSet
}

func NewBaseCommand(ctx clictx.Context, opts ...options.Options) BaseCommand {
	return BaseCommand{Context: ctx, OptionSet: opts}
}

func (BaseCommand) Complete(args []string) error { return nil }

func SetupCommand(ocmcmd OCMCommand, names ...string) *cobra.Command {
	c := ocmcmd.ForName(names[0])
	if !strings.HasSuffix(c.Use, names[0]+" ") {
		c.Use = names[0] + " " + c.Use
	}
	c.Aliases = names[1:]
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if set, ok := ocmcmd.(options.OptionSetProvider); ok {
			set.AsOptionSet().ProcessOnOptions(options.CompleteOptionsWithCLIContext(ocmcmd))
		}
		err := ocmcmd.Complete(args)
		if err == nil {
			err = ocmcmd.Run()
		}
		if err != nil && ocmcmd.StdErr() != os.Stderr {
			out.Error(ocmcmd, err.Error())
		}
		return err
	}
	c.TraverseChildren = true
	if u, ok := ocmcmd.(options.Usage); ok {
		c.Long = c.Long + u.Usage()
	}
	ocmcmd.AddFlags(c.Flags())
	return c
}
