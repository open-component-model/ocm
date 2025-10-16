package ociblobaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/mime"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repository string, digest digest.Digest, size int64, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	media := eff.MediaType
	if media == "" {
		media = mime.MIME_OCTET
	}
	spec := access.New(repository, digest, media, size)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repository string, digest digest.Digest, size int64, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repository, digest, size, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repository string, digest digest.Digest, size int64, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repository, digest, size, opts...)
}
