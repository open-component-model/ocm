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

package ocireg

import (
	"os"
	"sync"

	"github.com/gardener/ocm/pkg/common/accessio"
	"github.com/gardener/ocm/pkg/oci/cpi"
)

type RepositoryCloser interface {
	Close(ca *Repository) error
}

type RepositoryCloserFunction func(ca *Repository) error

func (f RepositoryCloserFunction) Close(ca *Repository) error {
	return f(ca)
}

////////////////////////////////////////////////////////////////////////////////

type RepositoryFormatHandler interface {
	CTFOption

	Format() accessio.FileFormat

	Open(ctx cpi.Context, path string, opts CTFOptions) (*Repository, error)
	Create(ctx cpi.Context, path string, opts CTFOptions, mode os.FileMode) (*Repository, error)
	Write(ca *Repository, path string, opts CTFOptions, mode os.FileMode) error
}

////////////////////////////////////////////////////////////////////////////////

var repoFormats = map[accessio.FileFormat]RepositoryFormatHandler{}
var lock sync.RWMutex

func RegisterRepositoryFormat(f RepositoryFormatHandler) {
	lock.Lock()
	defer lock.Unlock()
	repoFormats[f.Format()] = f
}

func GetRepositoryFormat(name accessio.FileFormat) RepositoryFormatHandler {
	lock.RLock()
	defer lock.RUnlock()
	return repoFormats[name]
}
