// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package formatoption

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/pkg/options"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/utils"
)

func From(o options.OptionSetProvider) *Option {
	var opt *Option
	o.AsOptionSet().Get(&opt)
	return opt
}

func New(list ...string) *Option {
	if len(list) > 0 {
		return &Option{List: list, Default: accessio.FileFormat(list[0])}
	}
	return &Option{Default: accessio.FormatDirectory}
}

type Option struct {
	format  string
	List    []string
	Default accessio.FileFormat
	Format  accessio.FileFormat

	flag *pflag.Flag
}

func (o *Option) setDefault() {
	if o.List == nil {
		o.List = accessio.GetFormats()
	}
	if o.Default == "" {
		o.Default = accessio.FormatDirectory
	}
}

func (o *Option) AddFlags(fs *pflag.FlagSet) {
	o.setDefault()
	fs.StringVarP(&o.format, "type", "t", string(o.Default), fmt.Sprintf("archive format (%s)", strings.Join(o.List, ", ")))
	o.flag = fs.Lookup("type")
}

func (o *Option) Configure(ctx clictx.Context) error {
	o.Format = accessio.FileFormat(o.format)
	for _, f := range o.List {
		if f == string(o.Format) {
			return nil
		}
	}
	return accessio.ErrInvalidFileFormat(o.format)
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
	return s + "\nThe default format is <code>directory</code>.\n"
}

func (o *Option) IsChanged() bool {
	return o.flag != nil && o.flag.Changed
}

func (o *Option) ChangedFormat() accessio.FileFormat {
	if o.IsChanged() {
		return o.Format
	} else {
		return ""
	}
}

func (o *Option) Mode() vfs.FileMode {
	mode := vfs.FileMode(0o660)
	if o.Format == accessio.FormatDirectory {
		mode = 0o770
	}
	return mode
}

var _ options.Options = (*Option)(nil)
