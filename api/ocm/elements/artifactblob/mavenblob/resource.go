package mavenblob

import (
	"github.com/mandelsoft/goutils/generics"
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/blobaccess/maven"
)

const TYPE = resourcetypes.MAVEN_PACKAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo *maven.Repository, groupId, artifactId, version string, opts ...Option) cpi.ArtifactAccess[M] {
	var eff Options
	WithCredentialContext(ctx).ApplyTo(&eff)
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTo(&eff)
		}
	}
	if eff.Blob.IsPackage() && eff.Hint == "" {
		eff.Hint = maven.NewCoordinates(groupId, artifactId, version).GAV()
	}

	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	blobprov := maven.Provider(repo, groupId, artifactId, version, &eff.Blob)
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
