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

package ctf

import (
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

func NewRepositorySpec(acc accessobj.AccessMode, path string, opts ...accessio.Option) *genericocireg.RepositorySpec {
	spec := ctf.NewRepositorySpec(acc, path, opts...)
	return genericocireg.NewRepositorySpec(spec, nil)
}

func Open(ctx cpi.Context, acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessio.Option) (cpi.Repository, error) {
	r, err := ctf.Open(ctx.OCIContext(), acc, path, mode, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepository(ctx, nil, r)
}

func Create(ctx cpi.Context, acc accessobj.AccessMode, path string, mode vfs.FileMode, opts ...accessio.Option) (cpi.Repository, error) {
	r, err := ctf.Create(ctx.OCIContext(), acc, path, mode, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepository(ctx, nil, r)
}
