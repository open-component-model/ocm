package mvnaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/mvn"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const TYPE = resourcetypes.MVN_ARTIFACT

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repository, groupId, artifactId, version string) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(repository, groupId, artifactId, version)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repository, groupId, artifactId, version string) cpi.ResourceAccess {
	return Access(ctx, meta, repository, groupId, artifactId, version)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repository, groupId, artifactId, version string) cpi.SourceAccess {
	return Access(ctx, meta, repository, groupId, artifactId, version)
}
