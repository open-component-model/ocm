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

package destoption

import (
	"fmt"

	"github.com/open-component-model/ocm/cmds/ocm/clictx"
	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

type Option struct {
	Destination    string
	PathFilesystem vfs.FileSystem
}

func (d *Option) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&d.Destination, "outfile", "O", "", "output file or directory")
}

func (o *Option) Complete(ctx clictx.Context) error {
	if o.Destination == "" {
		return fmt.Errorf("output destination required")
	}
	o.PathFilesystem = ctx.FileSystem()
	return nil
}

var _ options.Options = (*Option)(nil)
