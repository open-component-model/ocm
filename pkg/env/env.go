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

package env

import (
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/readonlyfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
)

////////////////////////////////////////////////////////////////////////////////

type Option interface {
	Mount(fs *composefs.ComposedFileSystem) error
}

type dummyOption struct{}

func (dummyOption) Mount(*composefs.ComposedFileSystem) error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type tdOpt struct {
	dummyOption
	path   string
	source string
}

func TestData(paths ...string) tdOpt {
	path := "/testdata"
	source := "testdata"

	switch len(paths) {
	case 0:
	case 1:
		source = paths[0]
	case 2:
		source = paths[0]
		path = paths[1]
	default:
		panic("invalid number of arguments")
	}
	return tdOpt{
		path:   path,
		source: source,
	}
}

func (o tdOpt) Mount(cfs *composefs.ComposedFileSystem) error {
	fs, err := projectionfs.New(osfs.New(), o.source)
	if err != nil {
		return err
	}
	fs = readonlyfs.New(fs)
	err = cfs.MkdirAll(o.path, vfs.ModePerm)
	if err != nil {
		return err
	}
	return cfs.Mount(o.path, fs)
}

////////////////////////////////////////////////////////////////////////////////

type Environment struct {
	vfs.VFS
	ctx        ocm.Context
	filesystem *composefs.ComposedFileSystem
}

func NewEnvironment(opts ...Option) *Environment {
	tmpfs, err := osfs.NewTempFileSystem()
	if err != nil {
		panic(err)
	}
	defer func() {
		vfs.Cleanup(tmpfs)
	}()
	err = tmpfs.Mkdir("/tmp", vfs.ModePerm)
	if err != nil {
		panic(err)
	}
	fs := composefs.New(tmpfs, "/tmp")
	for _, o := range opts {
		err := o.Mount(fs)
		if err != nil {
			panic(err)
		}
	}
	ctx := ocm.WithCredentials(credentials.WithConfigs(config.New()).New()).New()
	vfsattr.Set(ctx.AttributesContext(), fs)
	tmpfs = nil
	return &Environment{
		VFS:        vfs.New(fs),
		ctx:        ctx,
		filesystem: fs,
	}
}

var _ accessio.Option = (*Environment)(nil)

func (e *Environment) ApplyOption(options *accessio.Options) {
	options.PathFileSystem = e.FileSystem()
}

func (e *Environment) OCMContext() ocm.Context {
	return e.ctx
}

func (e *Environment) OCIContext() oci.Context {
	return e.ctx.OCIContext()
}

func (e *Environment) CredentialsContext() credentials.Context {
	return e.ctx.CredentialsContext()
}

func (e *Environment) ConfigContext() config.Context {
	return e.ctx.ConfigContext()
}

func (e *Environment) FileSystem() vfs.FileSystem {
	return vfsattr.Get(e.ctx)
}
