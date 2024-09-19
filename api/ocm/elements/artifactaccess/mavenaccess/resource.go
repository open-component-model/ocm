package mavenaccess

import (
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/maven"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/tech/maven"
)

const TYPE = resourcetypes.MAVEN_PACKAGE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repoUrl, groupId, artifactId, version, hint string, opts ...Option) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(repoUrl, groupId, artifactId, version, hint, opts...)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repoUrl, groupId, artifactId, version, hint string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repoUrl, groupId, artifactId, version, hint, opts...)
}

func ResourceAccessForMavenCoords(ctx ocm.Context, meta *cpi.ResourceMeta, repoUrl string, coords *maven.Coordinates) cpi.ResourceAccess {
	return Access(ctx, meta, repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, "", WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repoUrl, groupId, artifactId, version, hint string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repoUrl, groupId, artifactId, version, hint, opts...)
}

func SourceAccessForMavenCoords(ctx ocm.Context, meta *cpi.SourceMeta, repoUrl string, coords *maven.Coordinates) cpi.SourceAccess {
	return Access(ctx, meta, repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, "", WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))
}
