package ociartifactblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/api/oci"
	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
	blob "github.com/open-component-model/ocm/api/utils/blobaccess/ociartifact"
)

const TYPE = resourcetypes.OCI_IMAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, refname string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(append(opts, WithContext(ctx))...)
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
