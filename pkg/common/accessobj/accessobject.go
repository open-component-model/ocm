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

package accessobj

import (
	"fmt"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/errors"
)

type DescriptorHandlerFactory func(system vfs.FileSystem) StateHandler

////////////////////////////////////////////////////////////////////////////////

type AccessObjectInfo struct {
	DescriptorFileName       string
	ObjectTypeName           string
	ElementDirectoryName     string
	ElementTypeName          string
	DescriptorHandlerFactory DescriptorHandlerFactory
}

func (i *AccessObjectInfo) SubPath(name string) string {
	return filepath.Join(i.ElementDirectoryName, name)
}

// AccessObject provides a basic functionality for descriptor based access objects
// using a virtual filesystem for the internal representation.
type AccessObject struct {
	info   *AccessObjectInfo
	fs     vfs.FileSystem
	mode   vfs.FileMode
	state  State
	closer Closer
}

func NewAccessObject(info *AccessObjectInfo, acc AccessMode, fs vfs.FileSystem, setup Setup, closer Closer, mode vfs.FileMode) (*AccessObject, error) {
	defaulted, fs, err := InternalRepresentationFilesystem(acc, fs, info.ElementDirectoryName, mode)
	if err != nil {
		return nil, err
	}
	if setup != nil {
		err = setup.Setup(fs)
		if err != nil {
			return nil, err
		}
	}
	if defaulted {
		closer = FSCloser(closer)
	}

	s, err := NewFileBasedState(acc, fs, info.DescriptorFileName, "", info.DescriptorHandlerFactory(fs), mode)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	obj := &AccessObject{
		info:   info,
		state:  s,
		fs:     fs,
		mode:   mode,
		closer: closer,
	}

	return obj, nil
}

func (a *AccessObject) GetInfo() *AccessObjectInfo {
	return a.info
}

func (a *AccessObject) GetFileSystem() vfs.FileSystem {
	return a.fs
}

func (a *AccessObject) GetMode() vfs.FileMode {
	return a.mode
}

func (a *AccessObject) GetState() State {
	return a.state
}

func (a *AccessObject) IsClosed() bool {
	return a.fs == nil
}

func (a *AccessObject) IsReadOnly() bool {
	return a.state.IsReadOnly()
}

func (a *AccessObject) updateDescriptor() (bool, error) {
	if a.IsClosed() {
		return false, accessio.ErrClosed
	}
	return a.state.Update()
}

func (a *AccessObject) Write(path string, mode vfs.FileMode, opts ...accessio.Option) error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}

	o := accessio.AccessOptions(opts...)

	f := GetFormat(*o.FileFormat)
	if f == nil {
		return errors.ErrUnknown("file format", string(*o.FileFormat))
	}

	return f.Write(a, path, o, mode)
}

func (a *AccessObject) Update() error {
	if _, err := a.updateDescriptor(); err != nil {
		return fmt.Errorf("unable to update descriptor: %w", err)
	}

	return nil
}

func (a *AccessObject) Close() error {
	if a.IsClosed() {
		return accessio.ErrClosed
	}
	list := errors.ErrListf("close")
	list.Add(a.Update())
	if a.closer != nil {
		list.Add(a.closer.Close(a))
	}
	a.fs = nil
	return list.Result()
}
