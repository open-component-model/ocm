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
	"fmt"
	"os"

	"github.com/gardener/ocm/cmds/ocm/cmd"
	"github.com/gardener/ocm/cmds/ocm/commands/ocm"
	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/common/accessobj"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/gardener/ocm/pkg/ocm/compdesc"
	metav1 "github.com/gardener/ocm/pkg/ocm/compdesc/meta/v1"
	"github.com/gardener/ocm/pkg/ocm/repositories/ctf/comparch"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Options struct {
	Context cmd.Context

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
func NewCommand(ctx cmd.Context) *cobra.Command {
	opts := &Options{Context: ctx}
	cmd := &cobra.Command{
		Use:              "componentarchive [<options>] <component> <version> <provider> <path> {<label>=<value>}",
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(4),
		Aliases:          []string{"ca", "comparch"},
		Short:            "create new component archive",
		Long: `
create a new component archive. This might be either a directory prepared
to host component versuon content or a tar/tgz file.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.Run(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())
	return cmd
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.format, "type", "t", string(accessio.FormatDirectory), "archive format")
	fs.BoolVarP(&o.Force, "force", "f", false, "remove existing content")
}

func (o *Options) Complete(args []string) error {
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
		o.Labels, err = ocm.AddParsedLabel(o.Labels, a)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Options) Run() error {
	mode := vfs.FileMode(0660)
	if o.format == string(accessio.FormatDirectory) {
		mode = 0770
	}
	if ok, err := vfs.Exists(osfs.New(), o.Path); ok || err != nil {
		if err != nil {
			return err
		}
		if o.Force {
			err = os.RemoveAll(o.Path)
			if err != nil {
				return errors.Wrapf(err, "cannot remove old %q", o.Path)
			}
		}
	}
	obj, err := comparch.Create(o.Context.OCMContext(), accessobj.ACC_CREATE, o.Path, mode, o.Handler)
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
