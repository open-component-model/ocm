// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc

import (
	"reflect"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/equivalent"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
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
		equivalentElems(cd.Resources, o.Resources),
		equivalentElems(cd.Sources, o.Sources),
		equivalentElems(cd.References, o.References),
		equivalentSignatures(cd.Signatures, o.Signatures),
	)
}

func equivalentSignatures(a metav1.Signatures, b metav1.Signatures) equivalent.EqualState {
	if len(a) != len(b) {
		return equivalent.StateNotLocalHashEqual()
	}
outer:
	for _, s := range a {
		if o := b.GetByName(s.Name); o != nil {
			if reflect.DeepEqual(s, *o) {
				continue outer
			}
		}
		return equivalent.StateNotLocalHashEqual()
	}
	return equivalent.StateEquivalent()
}

func equivalentElems(a ElementAccessor, b ElementAccessor) equivalent.EqualState {
	state := equivalent.StateEquivalent()

	// Equivaluent of elements handles nil to provide state accoding to it
	// relevance for the signature.
	for i := 0; i < a.Len(); i++ {
		ea := a.Get(i)
		state = state.Apply(ea.Equivalent(GetByIdentity(b, ea.GetMeta().GetIdentity(a))))
	}
	for i := 0; i < b.Len(); i++ {
		eb := b.Get(i)
		if ea := GetByIdentity(a, eb.GetMeta().GetIdentity(b)); ea == nil {
			state = state.Apply(eb.Equivalent(ea))
		}
	}
	return state
}
