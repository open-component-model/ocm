package rscsel

import (
	"runtime"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extraid"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/ocm/selectors/accessors"
)

type (
	Selector     = selectors.ResourceSelector
	SelectorFunc = selectors.ResourceSelectorFunc
)

////////////////////////////////////////////////////////////////////////////////

type Relation v1.ResourceRelation

func (r Relation) MatchResource(list accessors.ElementListAccessor, res accessors.ResourceAccessor) bool {
	return v1.ResourceRelation(r) == res.GetRelation()
}

var (
	Local    = Relation(v1.LocalRelation)
	External = Relation(v1.ExternalRelation)
)

////////////////////////////////////////////////////////////////////////////////

func Executable(name string) Selector {
	return SelectorFunc(func(list accessors.ElementListAccessor, a accessors.ResourceAccessor) bool {
		extra := a.GetMeta().GetExtraIdentity()
		return a.GetMeta().GetName() == name && a.GetType() == resourcetypes.EXECUTABLE && extra != nil &&
			extra[extraid.ExecutableOperatingSystem] == runtime.GOOS &&
			extra[extraid.ExecutableArchitecture] == runtime.GOARCH
	})
}
