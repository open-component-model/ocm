package compdesc

import (
	"reflect"

	"ocm.software/ocm/api/ocm/compdesc/equivalent"
)

func (cd *ComponentDescriptor) Equal(obj interface{}) bool {
	if o, ok := obj.(*ComponentDescriptor); ok {
		if !cd.ObjectMeta.Equal(&o.ObjectMeta) {
			return false
		}
		if !reflect.DeepEqual(cd.Sources, o.Sources) {
			return false
		}
		if !reflect.DeepEqual(cd.Resources, o.Resources) {
			return false
		}
		if !reflect.DeepEqual(cd.References, o.References) {
			return false
		}
		if !reflect.DeepEqual(cd.Signatures, o.Signatures) {
			return false
		}
		if !reflect.DeepEqual(cd.NestedDigests, o.NestedDigests) {
			return false
		}
		return true
	}
	return false
}

func (cd *ComponentDescriptor) Equivalent(o *ComponentDescriptor) equivalent.EqualState {
	return equivalent.StateEquivalent().Apply(
		cd.ObjectMeta.Equivalent(o.ObjectMeta),
		cd.Resources.Equivalent(o.Resources),
		cd.Sources.Equivalent(o.Sources),
		cd.References.Equivalent(o.References),
		cd.Signatures.Equivalent(o.Signatures),
	)
}

func EquivalentElems(a ElementListAccessor, b ElementListAccessor) equivalent.EqualState {
	state := equivalent.StateEquivalent()

	// Equivalent of elements handles nil to provide state according to it
	// relevance for the signature.
	for i := 0; i < a.Len(); i++ {
		ea := a.Get(i)

		ib := GetIndexByIdentity(b, ea.GetMeta().GetIdentity(a))
		if ib != i {
			state = state.NotLocalHashEqual()
		}

		var eb ElementMetaAccessor
		if ib >= 0 {
			eb = b.Get(ib)
		}
		state = state.Apply(ea.Equivalent(eb))
	}
	for i := 0; i < b.Len(); i++ {
		eb := b.Get(i)
		if ea := GetByIdentity(a, eb.GetMeta().GetIdentity(b)); ea == nil {
			state = state.Apply(eb.Equivalent(ea))
		}
	}
	return state
}
