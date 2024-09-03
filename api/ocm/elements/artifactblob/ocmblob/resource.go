package ocmblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	base "ocm.software/ocm/api/utils/blobaccess/ocm"
)

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx cpi.Context, meta P, cvp base.ComponentVersionProvider, res base.ResourceProvider, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	blobprov := base.Provider(cvp, res)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx cpi.Context, meta *cpi.ResourceMeta, cvp base.ComponentVersionProvider, res base.ResourceProvider, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, cvp, res, opts...)
}

func SourceAccess(ctx cpi.Context, meta *cpi.SourceMeta, cvp base.ComponentVersionProvider, res base.ResourceProvider, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, cvp, res, opts...)
}
