package githubaccess

import (
	"github.com/mandelsoft/goutils/optionutils"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
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
