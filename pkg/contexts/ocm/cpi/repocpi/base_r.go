// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package repocpi

import (
	"io"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/refmgmt/resource"
)

type RepositoryImpl interface {
	resource.ResourceImplementation[cpi.Repository]
	internal.RepositoryImpl
}

type _RepositoryImplBase = resource.ResourceImplBase[cpi.Repository]

type RepositoryImplBase struct {
	*_RepositoryImplBase
	ctx cpi.Context
}

func (b *RepositoryImplBase) GetContext() cpi.Context {
	return b.ctx
}

func NewRepositoryImplBase(ctx cpi.Context, closer ...io.Closer) *RepositoryImplBase {
	base, _ := resource.NewResourceImplBase[cpi.Repository, io.Closer](nil, closer...)
	return &RepositoryImplBase{
		_RepositoryImplBase: base,
		ctx:                 ctx,
	}
}
