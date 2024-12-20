package cpi

import (
	"fmt"

	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/sliceutils"
	metav1 "ocm.software/ocm/api/ocm/refhints"
	"ocm.software/ocm/api/utils/runtime"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	cpi "ocm.software/ocm/api/ocm/internal"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
	ocm "ocm.software/ocm/api/ocm/types"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

////////////////////////////////////////////////////////////////////////////////

// ComponentVersionProvider should be implemented
// by Accesses based on component version instances.
// It is used to determine access type specific
// information. For example, OCI based access types
// may provide global OCI artifact references.
type ComponentVersionProvider interface {
	GetComponentVersion() (ComponentVersionAccess, error)
}

type ComponentVersionBasedAccessProvider struct {
	vers   ComponentVersionAccess
	access compdesc.AccessSpec
}

var (
	_ AccessProvider           = (*ComponentVersionBasedAccessProvider)(nil)
	_ ComponentVersionProvider = (*ComponentVersionBasedAccessProvider)(nil)
)

func NewBaseAccess(cv ComponentVersionAccess, acc compdesc.AccessSpec) *ComponentVersionBasedAccessProvider {
	return &ComponentVersionBasedAccessProvider{vers: cv, access: acc}
}

func (r *ComponentVersionBasedAccessProvider) GetOCMContext() Context {
	return r.vers.GetContext()
}

func (r *ComponentVersionBasedAccessProvider) GetComponentVersion() (ComponentVersionAccess, error) {
	return r.vers.Dup()
}

func (r *ComponentVersionBasedAccessProvider) ReferenceHintForAccess() metav1.ReferenceHints {
	if hp, ok := r.access.(cpi.HintProvider); ok {
		return hp.GetReferenceHint(r.vers)
	}
	return nil
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
	return m.AsBlobAccess(), nil
}

////////////////////////////////////////////////////////////////////////////////

type blobAccessProvider struct {
	ctx ocm.Context
	blobaccess.BlobAccessProvider
	hints  metav1.ReferenceHints
	global AccessSpec
}

var _ AccessProvider = (*blobAccessProvider)(nil)

func NewAccessProviderForBlobAccessProvider(ctx ocm.Context, prov blobaccess.BlobAccessProvider, hints []metav1.ReferenceHint, global AccessSpec) AccessProvider {
	return &blobAccessProvider{
		BlobAccessProvider: prov,
		hints:              hints,
		global:             global,
		ctx:                ctx,
	}
}

func (b *blobAccessProvider) GetOCMContext() cpi.Context {
	return b.ctx
}

func (b *blobAccessProvider) ReferenceHintForAccess() metav1.ReferenceHints {
	return b.hints
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

func NewArtifactAccessProviderForBlobAccessProvider[M any, P ReferenceHintProviderPointer[M]](ctx Context, meta *M, src blobAccessProvider, hint []metav1.ReferenceHint, global AccessSpec) cpi.ArtifactAccess[M] {
	return NewArtifactAccessForProvider[M, P](meta, NewAccessProviderForBlobAccessProvider(ctx, src, hint, global))
}

////////////////////////////////////////////////////////////////////////////////

type accessAccessProvider struct {
	ctx  ocm.Context
	spec AccessSpec
}

var _ AccessProvider = (*accessAccessProvider)(nil)

func NewAccessProviderForExternalAccessSpec(ctx ocm.Context, spec AccessSpec) (AccessProvider, error) {
	if spec.IsLocal(ctx) {
		return nil, fmt.Errorf("access spec describes a repository specific local access method")
	}
	return &accessAccessProvider{
		ctx:  ctx,
		spec: spec,
	}, nil
}

func (b *accessAccessProvider) GetOCMContext() cpi.Context {
	return b.ctx
}

func (b *accessAccessProvider) ReferenceHintForAccess() metav1.ReferenceHints {
	if h, ok := b.spec.(HintProvider); ok {
		return h.GetReferenceHint(&DummyComponentVersionAccess{b.ctx})
	}
	return nil
}

func (b *accessAccessProvider) GlobalAccess() cpi.AccessSpec {
	return nil
}

func (b *accessAccessProvider) Access() (cpi.AccessSpec, error) {
	return b.spec, nil
}

func (b *accessAccessProvider) AccessMethod() (cpi.AccessMethod, error) {
	return b.spec.AccessMethod(&DummyComponentVersionAccess{b.ctx})
}

func (b *accessAccessProvider) BlobAccess() (blobaccess.BlobAccess, error) {
	return accspeccpi.BlobAccessForAccessSpec(b.spec, &DummyComponentVersionAccess{b.ctx})
}

////////////////////////////////////////////////////////////////////////////////

type (
	accessProvider = AccessProvider
)

type ReferenceHintProviderPointer[P any] interface {
	compdesc.ReferenceHintProvider
	*P
}
type artifactAccessProvider[M any] struct {
	accessProvider
	componentVersionProvider ComponentVersionProvider
	meta                     *M
}

var _ credentials.ConsumerIdentityProvider = (*artifactAccessProvider[compdesc.ArtifactMetaAccess])(nil)

func NewArtifactAccessForProvider[M any, P ReferenceHintProviderPointer[M]](meta *M, prov AccessProvider) cpi.ArtifactAccess[M] {
	aa := &artifactAccessProvider[M]{
		accessProvider: prov,
		meta:           meta,
	}
	if p, ok := prov.(ComponentVersionProvider); ok {
		aa.componentVersionProvider = p
	}
	return aa
}

func (r *artifactAccessProvider[M]) Meta() *M {
	return r.meta
}

func (r *artifactAccessProvider[M]) GetReferenceHints() metav1.ReferenceHints {
	hints := any(r.meta).(compdesc.ReferenceHintProvider).GetReferenceHints()

	a, err := r.Access()
	if err == nil {
		cv, err := r.GetComponentVersion()
		if err == nil {
			defer cv.Close()
			sliceutils.AppendUniqueFunc(hints, runtime.MatchType[metav1.ReferenceHint], ReferenceHint(a, cv)...)
		}
	}
	return hints
}

func (b *artifactAccessProvider[M]) GetConsumerId(uctx ...credentials.UsageContext) credentials.ConsumerIdentity {
	m, err := b.AccessMethod()
	if err != nil {
		return nil
	}
	defer m.Close()
	return credentials.GetProvidedConsumerId(m, uctx...)
}

func (b *artifactAccessProvider[M]) GetIdentityMatcher() string {
	m, err := b.AccessMethod()
	if err != nil {
		return ""
	}
	defer m.Close()
	return credentials.GetProvidedIdentityMatcher(m)
}

func (b *artifactAccessProvider[M]) GetComponentVersion() (ComponentVersionAccess, error) {
	if b.componentVersionProvider != nil {
		return b.componentVersionProvider.GetComponentVersion()
	}
	return nil, nil
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
