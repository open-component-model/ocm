package npmblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	base "ocm.software/ocm/api/utils/blobaccess/npm"
	"ocm.software/ocm/api/utils/misc"
)

const TYPE = resourcetypes.NPM_PACKAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo, pkg, version string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(optionutils.WithDefaults(opts, WithHint(misc.NewNameVersion(pkg, version).String()), WithCredentialContext(ctx))...)

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
