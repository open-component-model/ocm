package mavenaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/maven"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/maven"
)

const TYPE = resourcetypes.MAVEN_ARTIFACT

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repoUrl, groupId, artifactId, version string, opts ...Option) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(repoUrl, groupId, artifactId, version, opts...)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repoUrl, groupId, artifactId, version string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repoUrl, groupId, artifactId, version, opts...)
}

func ResourceAccessForMavenCoords(ctx ocm.Context, meta *cpi.ResourceMeta, repoUrl string, coords *maven.Coordinates) cpi.ResourceAccess {
	return Access(ctx, meta, repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repoUrl, groupId, artifactId, version string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repoUrl, groupId, artifactId, version, opts...)
}

func SourceAccessForMavenCoords(ctx ocm.Context, meta *cpi.SourceMeta, repoUrl string, coords *maven.Coordinates) cpi.SourceAccess {
	return Access(ctx, meta, repoUrl, coords.GroupId, coords.ArtifactId, coords.Version, WithOptionalClassifier(coords.Classifier), WithOptionalExtension(coords.Extension))
}
