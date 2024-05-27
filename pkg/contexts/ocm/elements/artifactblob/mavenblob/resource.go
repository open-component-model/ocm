package mavenblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/open-component-model/ocm/pkg/blobaccess/maven"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/optionutils"
)

const TYPE = resourcetypes.MAVEN_ARTIFACT

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo *maven.Repository, groupId, artifactId, version string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(optionutils.WithDefaults(opts, WithCredentialContext(ctx))...)

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	blobprov := maven.BlobAccessProviderForMaven(repo, groupId, artifactId, version, &eff.Blob)
	accprov := cpi.NewAccessProviderForBlobAccessProvider(ctx, blobprov, eff.Hint, eff.Global)
	// strange type cast is required by Go compiler, meta has the correct type.
	return cpi.NewArtifactAccessForProvider(generics.Cast[*M](meta), accprov)
}

func ResourceAccess(ctx ocm.Context, meta *ocm.ResourceMeta, repo *maven.Repository, groupId, artifactId, version string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repo, groupId, artifactId, version, opts...)
}

func ResourceAccessForMavenCoords(ctx ocm.Context, meta *ocm.ResourceMeta, repo *maven.Repository, coords *maven.Coordinates, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))...)
}

func SourceAccess(ctx ocm.Context, meta *ocm.SourceMeta, repo *maven.Repository, groupId, artifactId, version string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repo, groupId, artifactId, version, opts...)
}

func SourceAccessForMavenCoords(ctx ocm.Context, meta *ocm.SourceMeta, repo *maven.Repository, coords *maven.Coordinates, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repo, coords.GroupId, coords.ArtifactId, coords.Version, optionutils.WithDefaults(opts, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))...)
}
