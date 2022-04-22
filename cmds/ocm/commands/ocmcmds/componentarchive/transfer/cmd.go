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
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/commands"
	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/names"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/utils"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/errors"
)

var (
	Names = names.ComponentArchive
	Verb  = commands.Transfer
)

type Command struct {
	utils.BaseCommand
	typ        string
	Path       string
	TargetName string
	FileFormat accessio.FileFormat
}

// NewCommand creates a new transfer command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{BaseCommand: utils.NewBaseCommand(ctx)}, names...)
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
	fs.StringVarP(&o.typ, "type", "t", "", "archive type to create (directory,tar,tgz)")
}

func (o *Command) Complete(args []string) error {
	o.Path = args[0]
	o.TargetName = args[1]

	if o.typ != "" {
		if accessobj.GetFormat(accessio.FileFormat(o.typ)) == nil {
			return errors.ErrInvalid(accessio.KIND_FILEFORMAT, o.typ)
		}
		o.FileFormat = accessio.FileFormat(o.typ)
	}
	return nil
}

func (o *Command) Run() error {
	session := ocm.NewSession(nil)
	defer session.Close()
	source, err := comparch.Open(o.Context.OCMContext(), accessobj.ACC_READONLY, o.Path, 0, o.Context)
	if err != nil {
		return err
	}
	session.Closer(source)

	target, ref, err := session.DetermineRepository(o.Context.OCMContext(), o.TargetName)
	if err != nil {
		if !errors.IsErrUnknown(err) || ref.Info == "" {
			return err
		}
		if ref.Type == "" {
			ref.Type = o.FileFormat.String()
		}
		if ref.Type == "" {
			return fmt.Errorf("ctf format type required to create ctf")
		}
		target, err = ctf.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, ref.Info, 0770, accessio.PathFileSystem(o.Context.FileSystem()))
		if err != nil {
			return err
		}
		session.Closer(target)
	}

	return ocm.TransferVersion(nil, source, target, nil)
}
