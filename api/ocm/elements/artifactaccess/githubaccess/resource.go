package githubaccess

import (
	"github.com/mandelsoft/goutils/optionutils"

	"github.com/open-component-model/ocm/api/ocm"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	"github.com/open-component-model/ocm/api/ocm/cpi"
	"github.com/open-component-model/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/github"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
)

const TYPE = resourcetypes.DIRECTORY_TREE

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, repo string, commit string, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(repo, eff.APIHostName, commit)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, repo string, commit string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, repo, commit, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, repo string, commit string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, repo, commit, opts...)
}
