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
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

// OCMCommand is a command pattern, that can be instantiated for a dediated
// sub command name to create a cobra command.
type OCMCommand interface {
	clictx.Context

	// ForName create a new cobra command for the given command name.
	// The Use attribute should omit the command name and just provide
	// a ost argument synopsis.
	// the complete attribute set is tweaked with the SetupCommand function
	// which calls this method.
	// Basically this should be an inherited function by the base implementation
	// but GO does not support virtual methods, therefore it is a global
	// function instead of a method.
	ForName(name string) *cobra.Command
	AddFlags(fs *pflag.FlagSet)
	Complete(args []string) error
	Run() error
}

// BaseCommand provides the basic functionality of an OCM command
// to carry a context and a set of reusable option specs.
type BaseCommand struct {
	clictx.Context
	options.OptionSet
}

func NewBaseCommand(ctx clictx.Context, opts ...options.Options) BaseCommand {
	return BaseCommand{Context: ctx, OptionSet: opts}
}

func (BaseCommand) Complete(args []string) error { return nil }

func addCommand(names []string, use string) string {
	if use == "" {
		return names[0]
	}
	lines := strings.Split(use, "\n")
outer:
	for i, l := range lines {
		if strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			continue
		}
		for _, n := range names {
			if strings.HasPrefix(l, n+" ") {
				continue outer
			}
		}
		lines[i] = names[0] + " " + l
	}
	return strings.Join(lines, "\n")
}

func HideCommand(cmd *cobra.Command) *cobra.Command {
	cmd.Hidden = true
	return cmd
}

func MassageCommand(cmd *cobra.Command, names ...string) *cobra.Command {
	cmd.Use = addCommand(names, cmd.Use)
	if len(names) > 1 {
		cmd.Aliases = names[1:]
	}
	cmd.DisableFlagsInUseLine = true
	cmd.TraverseChildren = true
	return cmd
}

// SetupCommand uses the OCMCommand to create and tweaks a cobra command
// to incorporate the additional reusable option specs and their usage documentation.
// Before the command executions the various Complete method flavors are
// executed on the additional options ond the OCMCommand.
func SetupCommand(ocmcmd OCMCommand, names ...string) *cobra.Command {
	c := ocmcmd.ForName(names[0])
	MassageCommand(c, names...)
	c.RunE = func(cmd *cobra.Command, args []string) error {
		var err error
		if set, ok := ocmcmd.(options.OptionSetProvider); ok {
			err = set.AsOptionSet().ProcessOnOptions(options.CompleteOptionsWithCLIContext(ocmcmd))
		}
		if err == nil {
			err = ocmcmd.Complete(args)
			if err == nil {
				err = ocmcmd.Run()
			}
		}
		/*
			if err != nil && ocmcmd.StdErr() != os.Stderr {
				out.Error(ocmcmd, err.Error())
			}
		*/
		return err
	}
	if u, ok := ocmcmd.(options.Usage); ok {
		c.Long += u.Usage()
	}
	ocmcmd.AddFlags(c.Flags())
	return c
}

func Names(def []string, names ...string) []string {
	if len(names) == 0 {
		return def
	}
	return names
}
