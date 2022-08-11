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

package formatoption

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/utils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(f ...accessio.FileFormat) *Option {
	if len(f) > 0 {
		return &Option{Default: f[0]}
	}
	return &Option{Default: accessio.FormatDirectory}
}

type Option struct {
	format  string
	Default accessio.FileFormat
	Format  accessio.FileFormat
}

func (o *Option) setDefault() {
	if o.Default == "" {
		o.Default = accessio.FormatDirectory
	}
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.setDefault()
	fs.StringVarP(&o.format, "type", "t", string(o.Default), "archive format")

}

func (o *Option) Complete(ctx clictx.Context) error {
	o.Format = accessio.FileFormat(o.format)
	if accessobj.GetFormat(o.Format) == nil {
		return accessio.ErrInvalidFileFormat(o.format)
	}
	return nil
}

func (o *Option) Usage() string {
	o.setDefault()
	s := `
The <code>--type</code> option accepts a file format for the
target archive to use. The following formats are supported:
`
	list := utils.StringSlice{}
	for k := range accessobj.GetFormats() {
		list.Add(k.String())
	}
	list.Sort()
	for _, k := range list {
		s = s + "- " + k + "\n"
	}
	return s + "The default format is <code>directory</code>.\n"
}

func (o *Option) Mode() vfs.FileMode {
	mode := vfs.FileMode(0660)
	if o.Format == accessio.FormatDirectory {
		mode = 0770
	}
	return mode
}

var _ options.Options = (*Option)(nil)
