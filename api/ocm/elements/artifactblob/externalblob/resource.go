package externalblob

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/refhints"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec, opts ...Option) (cpi.ArtifactAccess[M], error) {
	eff := optionutils.EvalOptions(opts...)

	hint := eff.Hint
	if len(hint) == 0 {
		hint = ocm.ReferenceHint(access, &cpi.DummyComponentVersionAccess{ctx})
	}
	global := eff.Global
	if global == nil {
		global = ocm.GlobalAccess(access, ctx)
	}

	prov, err := cpi.NewAccessProviderForExternalAccessSpec(ctx, access)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid external access method %q", access.GetKind())
	}
	return cpi.NewArtifactAccessForProvider[M, P](generics.Cast[*M](meta), newAccessProvider(prov, hint, global)), nil
}

type _accessProvider = cpi.AccessProvider

type accessProvider struct {
	_accessProvider
	hint   refhints.ReferenceHints
	global cpi.AccessSpec
}

func newAccessProvider(prov cpi.AccessProvider, hint refhints.ReferenceHints, global cpi.AccessSpec) cpi.AccessProvider {
	return &accessProvider{
		_accessProvider: prov,
		hint:            hint,
		global:          global,
	}
}

func (p *accessProvider) ReferenceHintForAccess() refhints.ReferenceHints {
	if len(p.hint) != 0 {
		return p.hint
	}
	return p._accessProvider.ReferenceHintForAccess()
}

func (p *accessProvider) GlobalAccess() cpi.AccessSpec {
	if p.global != nil {
		return p.global
	}
	return p._accessProvider.GlobalAccess()
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, access cpi.AccessSpec, opts ...Option) (cpi.ResourceAccess, error) {
	return Access(ctx, meta, access, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, access cpi.AccessSpec, opts ...Option) (cpi.SourceAccess, error) {
	return Access(ctx, meta, access, opts...)
}
