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

package spiff

import (
	"fmt"

	"github.com/mandelsoft/spiff/spiffing"
	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/errors"
)

type Option interface {
	ApplyToRequest(r *Request) error
}

type Options []Option

func (o *Options) Add(opt Option) *Options {
	if opt != nil {
		*o = append(*o, opt)
	}
	return o
}

func (o Options) ApplyToRequest(r *Request) error {
	for _, o := range o {
		if o != nil {
			err := o.ApplyToRequest(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetRequest(opts ...Option) (*Request, error) {
	req := &Request{}
	err := Options(opts).ApplyToRequest(req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type OptionFunction func(r *Request) error

func (f OptionFunction) ApplyToRequest(r *Request) error {
	return f(r)
}

func FileSystem(fs vfs.FileSystem) OptionFunction {
	return func(r *Request) error {
		r.FileSystem = accessio.FileSystem(fs)
		return nil
	}
}
func Context(ctx datacontext.Context) OptionFunction {
	return FileSystem(vfsattr.Get(ctx))
}
func Values(values interface{}) OptionFunction {
	return func(r *Request) error {
		r.Values = values
		return nil
	}
}
func ValuesNode(values string) OptionFunction {
	return func(r *Request) error {
		r.ValuesNode = values
		return nil
	}
}
func StubData(name string, data []byte) OptionFunction {
	return func(r *Request) error {
		if len(data) > 0 {
			r.Stubs = append(r.Stubs, spiffing.NewSourceData(name, data))
		}
		return nil
	}
}
func TemplateData(name string, data []byte) OptionFunction {
	return func(r *Request) error {
		if len(data) == 0 {
			return fmt.Errorf("no template data for " + name)
		}
		r.Template = spiffing.NewSourceData(name, data)
		return nil
	}
}
func StubFile(path string, fss ...vfs.FileSystem) OptionFunction {
	return func(r *Request) error {
		r.Stubs = append(r.Stubs, spiffing.NewSourceFile(path, accessio.FileSystem(append(fss, r.FileSystem)...)))
		return nil
	}
}
func TemplateFile(path string, fss ...vfs.FileSystem) OptionFunction {
	return func(r *Request) error {
		r.Template = spiffing.NewSourceFile(path, accessio.FileSystem(append(fss, r.FileSystem)...))
		return nil
	}
}
func WorkDir(path string) OptionFunction {
	return func(r *Request) error {
		fs, err := cwdfs.New(r.FileSystem, path)
		if err != nil {
			return errors.Wrapf(err, "cannot set working directory %s", path)
		}
		r.FileSystem = fs
		return nil
	}
}

func Validated(schemedata []byte, opts ...Option) Option {
	if schemedata == nil {
		return Options(opts)
	}
	return OptionFunction(func(r *Request) error {
		tmp := *r
		tmp.Template = nil
		tmp.Stubs = nil
		err := Options(opts).ApplyToRequest(&tmp)
		if err != nil {
			return err
		}
		if tmp.Template != nil {
			err = ValidateSourceByScheme(tmp.Template, schemedata)
			if err != nil {
				return errors.Wrapf(err, "template %s", tmp.Template.Name())
			}
		}
		for _, s := range tmp.Stubs {
			err = ValidateSourceByScheme(s, schemedata)
			if err != nil {
				return errors.Wrapf(err, "validating %s", s.Name())
			}
		}
		return Options(opts).ApplyToRequest(r)
	})
}
