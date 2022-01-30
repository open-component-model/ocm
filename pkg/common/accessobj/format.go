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
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/errors"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

type FormatHandler interface {
	Option

	Format() accessio.FileFormat

	Open(info *AccessObjectInfo, acc AccessMode, path string, opts Options) (*AccessObject, error)
	Create(info *AccessObjectInfo, path string, opts Options, mode vfs.FileMode) (*AccessObject, error)
	Write(obj *AccessObject, path string, opts Options, mode vfs.FileMode) error
}

////////////////////////////////////////////////////////////////////////////////

var fileFormats = map[accessio.FileFormat]FormatHandler{}
var lock sync.RWMutex

func RegisterFormat(f FormatHandler) {
	lock.Lock()
	defer lock.Unlock()
	fileFormats[f.Format()] = f
}

func GetFormat(name accessio.FileFormat) FormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return fileFormats[name]
}

////////////////////////////////////////////////////////////////////////////////

type Closer interface {
	Close(*AccessObject) error
}

type CloserFunction func(*AccessObject) error

func (f CloserFunction) Close(obj *AccessObject) error {
	return f(obj)
}

////////////////////////////////////////////////////////////////////////////////

type fsCloser struct {
	closer Closer
}

func FSCloser(closer Closer) Closer {
	return &fsCloser{closer}
}

func (f fsCloser) Close(obj *AccessObject) error {
	err := errors.ErrListf("cannot close %s", obj.info.ObjectTypeName)
	if f.closer != nil {
		err.Add(f.closer.Close(obj))
	}
	err.Add(vfs.Cleanup(obj.fs))
	return err.Result()
}
