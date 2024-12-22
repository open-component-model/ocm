package dockerdaemonblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/refhints"
	oci2 "ocm.software/ocm/api/tech/oci"
	"ocm.software/ocm/api/utils/blobaccess/dockerdaemon"
)

const TYPE = resourcetypes.OCI_IMAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, name string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}
	eff.Blob.Context = ctx.OCIContext()
	locator, version, err := dockerdaemon.ImageInfoFor(name, &eff.Blob)
	if err == nil {
		version = eff.Blob.Version
	}
	h := eff.Hint.GetReferenceHint(oci2.ReferenceHintType, "")
	hint := ociartifact.Hint(optionutils.AsValue(eff.Blob.Origin), locator, h.GetReference(), version)
	blobprov := dockerdaemon.Provider(name, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, refhints.ReferenceHints{oci2.ReferenceHint(hint)}, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider[M, P](generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, path string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, path, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, path string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, path, opts...)
}
