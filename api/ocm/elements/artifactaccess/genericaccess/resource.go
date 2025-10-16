package genericaccess

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec) (cpi.ArtifactAccess[M], error) {
	prov, err := cpi.NewAccessProviderForExternalAccessSpec(ctx, access)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid external access method %q", access.GetKind())
	}
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), prov), nil
}

func MustAccess[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, access ocm.AccessSpec) cpi.ArtifactAccess[M] {
	a, err := Access(ctx, meta, access)
	if err != nil {
		panic(err)
	}
	return a
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, access ocm.AccessSpec) (cpi.ResourceAccess, error) {
	return Access(ctx, meta, access)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, access ocm.AccessSpec) (cpi.SourceAccess, error) {
	return Access(ctx, meta, access)
}
