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

// ComponentAccessImpl is the provider implementation
// interface for component versions.
type ComponentAccessImpl interface {
	resource.ResourceImplementation[cpi.ComponentAccess]
	internal.ComponentAccessImpl

	IsReadOnly() bool
	GetName() string

	IsOwned(access cpi.ComponentVersionAccess) bool

	AddVersion(cv cpi.ComponentVersionAccess) error
}

type _ComponentAccessImplBase = resource.ResourceImplBase[cpi.ComponentAccess]

type ComponentAccessImplBase struct {
	*_ComponentAccessImplBase
	ctx  cpi.Context
	name string
}

func NewComponentAccessImplBase(ctx cpi.Context, name string, repo RepositoryViewManager, closer ...io.Closer) (*ComponentAccessImplBase, error) {
	base, err := resource.NewResourceImplBase[cpi.ComponentAccess](repo, closer...)
	if err != nil {
		return nil, err
	}
	return &ComponentAccessImplBase{
		_ComponentAccessImplBase: base,
		ctx:                      ctx,
		name:                     name,
	}, nil
}

func (b *ComponentAccessImplBase) GetContext() cpi.Context {
	return b.ctx
}

func (b *ComponentAccessImplBase) GetName() string {
	return b.name
}
