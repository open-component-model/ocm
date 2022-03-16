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

package testhelper

import (
	"github.com/gardener/ocm/cmds/ocm/app"
	"github.com/gardener/ocm/cmds/ocm/clictx"
	"github.com/gardener/ocm/pkg/ocm"
	"github.com/mandelsoft/vfs/pkg/composefs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/readonlyfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
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

type fsOpt struct {
	dummyOption
	path string
	fs   vfs.FileSystem
}

func FileSystem(fs vfs.FileSystem, path string) fsOpt {
	return fsOpt{
		path: path,
		fs:   fs,
	}
}

func (o fsOpt) Mount(cfs *composefs.ComposedFileSystem) error {
	return cfs.Mount(o.path, o.fs)
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

type TestEnv struct {
	vfs.VFS
	filesystem *composefs.ComposedFileSystem
	clictx.Context
	*app.CLI
}

func NewTestEnv(opts ...Option) *TestEnv {
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
	ctx := clictx.WithOCM(ocm.DefaultContext()).WithFileSystem(fs).New()
	tmpfs = nil
	return &TestEnv{
		VFS:        vfs.New(fs),
		filesystem: fs,
		Context:    ctx,
		CLI:        app.NewCLI(ctx),
	}
}
