package wgetblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/pkg/blobaccess/wget"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(optionutils.WithDefaults(opts, WithCredentialContext(ctx))...)

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
