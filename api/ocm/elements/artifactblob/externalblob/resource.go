package externalblob

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec, opts ...Option) (cpi.ArtifactAccess[M], error) {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}

	hint := eff.Hint
	if hint == "" {
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
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), newAccessProvider(prov, hint, global)), nil
}

type _accessProvider = cpi.AccessProvider

type accessProvider struct {
	_accessProvider
	hint   string
	global cpi.AccessSpec
}

func newAccessProvider(prov cpi.AccessProvider, hint string, global cpi.AccessSpec) cpi.AccessProvider {
	return &accessProvider{
		_accessProvider: prov,
		hint:            hint,
		global:          global,
	}
}

func (p *accessProvider) ReferenceHint() string {
	if p.hint != "" {
		return p.hint
	}
	return p._accessProvider.ReferenceHint()
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
