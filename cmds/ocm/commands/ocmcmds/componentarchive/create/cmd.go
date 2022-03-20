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

package create

import (
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/cmds/ocm/commands/ocmcmds"
	"github.com/gardener/ocm/cmds/ocm/pkg/utils"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/repositories/comparch/comparch"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Command struct {
	Context clictx.Context

	format string

	Handler comparch.FormatHandler
	Force   bool
	Path    string

	Component string
	Version   string
	Provider  string
	Labels    metav1.Labels
}

// NewCommand creates a new ctf command.
func NewCommand(ctx clictx.Context, names ...string) *cobra.Command {
	return utils.SetupCommand(&Command{Context: ctx}, names...)
}

func (o *Command) ForName(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "[<options>] <component> <version> <provider> <path> {<label>=<value>}",
		Args:  cobra.MinimumNArgs(4),
		Short: "create new component archive",
		Long: `
Create a new component archive. This might be either a directory prepared
to host component version content or a tar/tgz file.
`,
	}
}

func (o *Command) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.format, "type", "t", string(accessio.FormatDirectory), "archive format")
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
}

func (o *Command) Complete(args []string) error {
	var err error

	o.Handler = comparch.GetFormat(accessio.FileFormat(o.format))
	if o.Handler == nil {
		return accessio.ErrInvalidFileFormat(o.format)
	}

	o.Component = args[0]
	o.Version = args[1]
	o.Provider = args[2]
	o.Path = args[3]

	for _, a := range args[4:] {
		o.Labels, err = ocmcmds.AddParsedLabel(o.Labels, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Command) Run() error {
	mode := vfs.FileMode(0660)
	if o.format == string(accessio.FormatDirectory) {
		mode = 0770
	}
	fs := o.Context.FileSystem()
	if ok, err := vfs.Exists(fs, o.Path); ok || err != nil {
		if err != nil {
			return err
		}
		if o.Force {
			err = fs.RemoveAll(o.Path)
			if err != nil {
				return errors.Wrapf(err, "cannot remove old %q", o.Path)
			}
		}
	}
	obj, err := comparch.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, o.Path, mode, o.Handler, accessio.PathFileSystem(fs))
	if err != nil {
		return err
	}
	desc := obj.GetDescriptor()
	desc.Name = o.Component
	desc.Version = o.Version
	desc.Provider = metav1.ProviderType(o.Provider)
	desc.Labels = o.Labels

	err = compdesc.Validate(desc)
	if err != nil {
		obj.Close()
		return errors.Newf("invalid component info: %s", err)
	}
	return obj.Close()
}
