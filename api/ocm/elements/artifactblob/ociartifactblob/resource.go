package ociartifactblob

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	blob "ocm.software/ocm/api/utils/blobaccess/ociartifact"
)

const TYPE = resourcetypes.OCI_IMAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, refname string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	WithContext(ctx).ApplyTo(&eff)
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	hint := eff.Hint
	if hint == "" {
		ref, err := oci.ParseRef(refname)
		if err == nil {
			hint = ref.String()
		}
	}

	blobprov := blob.Provider(refname, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, path string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, path, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, path string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, path, opts...)
}
