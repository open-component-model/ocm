// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ctf

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
)

func NewRepositorySpec(acc accessobj.AccessMode, path string, opts ...accessio.Option) (*genericocireg.RepositorySpec, error) {
	spec, err := ctf.NewRepositorySpec(acc, path, opts...)
	if err != nil {
		return nil, err
	}
	return genericocireg.NewRepositorySpec(spec, nil), nil
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
