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

package artefactset

import (
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Options struct {
	accessio.StandardOptions

	FormatVersion string `json:"formatVersion,omitempty"`
}

func NewOptions(olist ...accessio.Option) (*Options, error) {
	opts := &Options{}
	err := accessio.ApplyOptions(opts, olist...)
	if err != nil {
		return nil, err
	}
	return opts, nil
}

type FormatVersionOption interface {
	SetFormatVersion(string)
	GetFormatVersion() string
}

func GetFormatVersion(opts accessio.Options) string {
	if o, ok := opts.(FormatVersionOption); ok {
		return o.GetFormatVersion()
	}
	return ""
}

var _ FormatVersionOption = (*Options)(nil)

func (o *Options) SetFormatVersion(s string) {
	o.FormatVersion = s
}

func (o *Options) GetFormatVersion() string {
	return o.FormatVersion
}

func (o *Options) ApplyOption(opts accessio.Options) error {
	err := o.StandardOptions.ApplyOption(opts)
	if err != nil {
		return err
	}
	if o.FormatVersion != "" {
		if s, ok := opts.(FormatVersionOption); ok {
			s.SetFormatVersion(o.FormatVersion)
		} else {
			return errors.ErrNotSupported("format version option")
		}
	}
	return nil
}

type optFmt struct {
	format string
}

var _ accessio.Option = (*optFmt)(nil)

func StructureFormat(fmt string) accessio.Option {
	return &optFmt{fmt}
}

func (o *optFmt) ApplyOption(opts accessio.Options) error {
	if s, ok := opts.(FormatVersionOption); ok {
		s.SetFormatVersion(o.format)
		return nil
	}
	return errors.ErrNotSupported("format version option")
}
