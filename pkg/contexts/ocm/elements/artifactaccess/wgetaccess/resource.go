package wgetaccess

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	access "github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/wget"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/elements/artifactaccess/genericaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
)

const TYPE = resourcetypes.BLOB

func Access[M any, P compdesc.ArtifactMetaPointer[M]](ctx ocm.Context, meta P, url string, opts ...Option) cpi.ArtifactAccess[M] {
	if meta.GetType() == "" {
		meta.SetType(TYPE)
	}

	spec := access.New(url, opts...)
	// is global access, must work, otherwise there is an error in the lib.
	return genericaccess.MustAccess(ctx, meta, spec)
}

func ResourceAccess(ctx ocm.Context, meta *cpi.ResourceMeta, url string, opts ...Option) cpi.ResourceAccess {
	return Access(ctx, meta, url, opts...)
}

func SourceAccess(ctx ocm.Context, meta *cpi.SourceMeta, url string, opts ...Option) cpi.SourceAccess {
	return Access(ctx, meta, url, opts...)
}
