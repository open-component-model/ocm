package wgetblob

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/blobaccess/wget"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	WithCredentialContext(ctx).ApplyTo(&eff)
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	blobprov := wget.Provider(url, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, url string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, url, opts...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, url string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, url, opts...)
}
