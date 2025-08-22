package dockermultiblob

import (
	"github.com/mandelsoft/goutils/generics"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/blobaccess/dockermulti"
)

const TYPE = resourcetypes.OCI_IMAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	eff.Blob.Context = ctx.OCIContext()

	blobprov := dockermulti.Provider(&eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, name string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, opts...)
}
