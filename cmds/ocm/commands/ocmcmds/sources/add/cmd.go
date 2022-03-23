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

package add

import (
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds/common"
	"github.com/gardener/ocm/cmds/ocm/pkg/template"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/spf13/cobra"
)

type Command struct {
	common.ResourceAdderCommand
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{common.ResourceAdderCommand{Context: ctx}}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <target> {<resourcefile> | <var>=<value>}",
		Args:  cobra.MinimumNArgs(2),
		Short: "add source information to a component version",
		Long: `
Add  source information specified in a resource file to a component version.
So far only component archives are supported as target.
` + (&template.Options{}).Usage(),
	}
}

func (o *Command) Run() error {
	return o.ProcessResourceDescriptions("sources", ResourceSpecHandler{})
}
