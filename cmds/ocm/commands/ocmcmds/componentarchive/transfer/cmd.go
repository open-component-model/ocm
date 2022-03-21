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

package transfer

import (
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/gardener/ocm/pkg/ocm/repositories/comparch/comparch"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Command struct {
	Context clictx.Context

	Path       string
	TargetName string

	Source ocm.ComponentVersionAccess
	Target ocm.Repository
}

// NewCommand creates a new transfer command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>]  <source> <target>",
		Args:  cobra.MinimumNArgs(2),
		Short: "transfer component archive to some component repository",
		Long: `
Transfer a component archive to some component repository. This might
be a CTF Archive or a regular repository.
Explicitly supported types, so far: OCIRegistry, CTF (directory, tar, tgz).
If the type CTF is specified the target must already exist, if CTF flavor
is specified it will be created if it does not exist.

Besides those explicitly known types a complete repository spec might be configured,
either via inline argument or command configuration file and name.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
}

func (o *Command) Complete(args []string) error {
	var err error

	o.Path = args[0]
	o.TargetName = args[1]

	o.Source, err = comparch.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.Path, 0, o.Context)
	if err != nil {
		return err
	}

	o.Target, err = o.Context.OCM().DetermineRepository(o.TargetName)
	if err != nil {
		o.Source.Close()
		return err
	}

	return nil
}

func (o *Command) Run() error {
	defer o.Source.Close()
	defer o.Target.Close()
	return ocm.TransferVersion(nil, o.Source, o.Target, nil)
}
