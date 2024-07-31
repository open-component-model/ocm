package rscsel

import (
	"runtime"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/extraid"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors/accessors"
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
