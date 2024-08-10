package githubaccess

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/elements/artifactaccess/genericaccess"
	access "ocm.software/ocm/api/ocm/extensions/accessmethods/git"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, opts ...Option) cpi.ArtifactAccess[M] {
	eff := optionutils.EvalOptions(opts...)
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(eff.URL, eff.Ref, eff.PathSpec)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}
