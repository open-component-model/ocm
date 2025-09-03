package npmblob

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	base "ocm.software/ocm/api/utils/blobaccess/npm"
	common "ocm.software/ocm/api/utils/misc"
)

const TYPE = resourcetypes.NPM_PACKAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo, pkg, version string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	WithHint(common.NewNameVersion(pkg, version).String()).ApplyTo(&eff)
	WithCredentialContext(ctx).ApplyTo(&eff)
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	blobprov := base.Provider(repo, pkg, version, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, repo, pkg, version string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repo, pkg, version, opts...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, repo, pkg, version string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repo, pkg, version, opts...)
}
