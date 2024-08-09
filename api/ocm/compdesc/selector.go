package compdesc

import (
	"github.com/mandelsoft/goutils/generics"

	"ocm.software/ocm/api/ocm/selectors/accessors"
)

type elemList struct {
	ElementAccessor
}

func (e *elemList) Get(i int) accessors.ElementMetaAccessor {
	return generics.Cast[accessors.ElementMetaAccessor](e.ElementAccessor.Get(i))
}

func MapToSelectorElementList(accessor ElementAccessor) accessors.ElementListAccessor {
	return &elemList{accessor}
}

////////////////////////////////////////////////////////////////////////////////

type rscAcc struct {
	*Resource
}

func (a rscAcc) GetMeta() accessors.ElementMeta {
	return a.Resource.GetMeta()
}

func MapToSelectorResource(r *Resource) accessors.ResourceAccessor {
	return rscAcc{r}
}

////////////////////////////////////////////////////////////////////////////////

type srcAcc struct {
	*Source
}

func (a srcAcc) GetMeta() accessors.ElementMeta {
	return a.Source.GetMeta()
}

func MapToSelectorSource(r *Source) accessors.SourceAccessor {
	return srcAcc{r}
}

////////////////////////////////////////////////////////////////////////////////

type refAcc struct {
	*Reference
}

func (a refAcc) GetMeta() accessors.ElementMeta {
	return a.Reference.GetMeta()
}

func MapToSelectorReference(r *Reference) accessors.ReferenceAccessor {
	return refAcc{r}
}
