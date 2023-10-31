// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi

import (
	"github.com/open-component-model/ocm/pkg/common/accessio/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/context"
	cpi "github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/descriptor"
	"github.com/open-component-model/ocm/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////

type ComponentVersionBasedAccessProvider struct {
	vers   ComponentVersionAccess
	access compdesc.AccessSpec
}

var _ AccessProvider = (*ComponentVersionBasedAccessProvider)(nil)

// Deprecated: use ComponentVersionBasedAccessProvider.
type BaseAccess = ComponentVersionBasedAccessProvider

func NewBaseAccess(cv ComponentVersionAccess, acc compdesc.AccessSpec) *ComponentVersionBasedAccessProvider {
	return &ComponentVersionBasedAccessProvider{vers: cv, access: acc}
}

func (r *ComponentVersionBasedAccessProvider) GetOCMContext() Context {
	return r.vers.GetContext()
}

func (r *ComponentVersionBasedAccessProvider) ReferenceHint() string {
	if hp, ok := r.access.(cpi.HintProvider); ok {
		return hp.GetReferenceHint(r.vers)
	}
	return ""
}

func (r *ComponentVersionBasedAccessProvider) GlobalAccess() AccessSpec {
	acc, err := r.GetOCMContext().AccessSpecForSpec(r.access)
	if err != nil {
		return nil
	}
	return acc.GlobalAccessSpec(r.GetOCMContext())
}

func (r *ComponentVersionBasedAccessProvider) Access() (AccessSpec, error) {
	return r.vers.GetContext().AccessSpecForSpec(r.access)
}

func (r *ComponentVersionBasedAccessProvider) AccessMethod() (AccessMethod, error) {
	acc, err := r.vers.GetContext().AccessSpecForSpec(r.access)
	if err != nil {
		return nil, err
	}
	return acc.AccessMethod(r.vers)
}

func (r *ComponentVersionBasedAccessProvider) BlobAccess() (BlobAccess, error) {
	m, err := r.AccessMethod()
	if err != nil {
		return nil, err
	}
	return BlobAccessForAccessMethod(AccessMethodAsView(m))
}

////////////////////////////////////////////////////////////////////////////////

type blobAccessProvider struct {
	ctx ocm.Context
	blobaccess.BlobAccessProvider
	hint   string
	global AccessSpec
}

var _ AccessProvider = (*blobAccessProvider)(nil)

func NewAccessProviderForBlobAccessProvider(ctx ocm.Context, prov blobaccess.BlobAccessProvider, hint string, global AccessSpec) AccessProvider {
	return &blobAccessProvider{
		BlobAccessProvider: prov,
		hint:               hint,
		global:             global,
		ctx:                ctx,
	}
}

func (b *blobAccessProvider) GetOCMContext() cpi.Context {
	return b.ctx
}

func (b *blobAccessProvider) ReferenceHint() string {
	return b.hint
}

func (b *blobAccessProvider) GlobalAccess() cpi.AccessSpec {
	return b.global
}

func (b blobAccessProvider) Access() (cpi.AccessSpec, error) {
	return nil, errors.ErrNotFound(descriptor.KIND_ACCESSMETHOD)
}

func (b *blobAccessProvider) AccessMethod() (cpi.AccessMethod, error) {
	return nil, errors.ErrNotFound(descriptor.KIND_ACCESSMETHOD)
}

////////////////////////////////////////////////////////////////////////////////

func NewArtifactAccessProviderForBlobAccessProvider[M any](ctx Context, meta *M, src blobAccessProvider, hint string, global AccessSpec) cpi.ArtifactAccess[M] {
	return NewArtifactAccessForProvider(meta, NewAccessProviderForBlobAccessProvider(ctx, src, hint, global))
}

////////////////////////////////////////////////////////////////////////////////

type accessProvider = AccessProvider

type artifactAccessProvider[M any] struct {
	accessProvider
	meta *M
}

func NewArtifactAccessForProvider[M any](meta *M, prov AccessProvider) cpi.ArtifactAccess[M] {
	return &artifactAccessProvider[M]{
		accessProvider: prov,
		meta:           meta,
	}
}

func (r *artifactAccessProvider[M]) Meta() *M {
	return r.meta
}

////////////////////////////////////////////////////////////////////////////////

var _ ResourceAccess = (*artifactAccessProvider[ResourceMeta])(nil)

func NewResourceAccess(componentVersion ComponentVersionAccess, accessSpec compdesc.AccessSpec, meta ResourceMeta) ResourceAccess {
	return NewResourceAccessForProvider(&meta, NewBaseAccess(componentVersion, accessSpec))
}

func NewResourceAccessForProvider(meta *ResourceMeta, prov AccessProvider) ResourceAccess {
	return NewArtifactAccessForProvider(meta, prov)
}

////////////////////////////////////////////////////////////////////////////////

var _ SourceAccess = (*artifactAccessProvider[SourceMeta])(nil)

func NewSourceAccess(componentVersion ComponentVersionAccess, accessSpec compdesc.AccessSpec, meta SourceMeta) SourceAccess {
	return NewSourceAccessForProvider(&meta, NewBaseAccess(componentVersion, accessSpec))
}

func NewSourceAccessForProvider(meta *SourceMeta, prov AccessProvider) SourceAccess {
	return NewArtifactAccessForProvider(meta, prov)
}
