// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc

import (
	"reflect"
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

func (cd *ComponentDescriptor) Equivalent(o *ComponentDescriptor) (equal bool, detectable bool) {
	if !cd.ObjectMeta.Equal(&o.ObjectMeta) {
		return false, true
	}

	if e, d := equivalentElems(cd.Resources, o.Resources); !e {
		return e, d
	}
	if e, d := equivalentElems(cd.Sources, o.Sources); !e {
		return e, d
	}
	if e, d := equivalentElems(cd.References, o.References); !e {
		return e, d
	}
	return true, true
}

func equivalentElems(a ElementAccessor, b ElementAccessor) (equal bool, detectable bool) {
	if a.Len() != b.Len() {
		return false, true
	}

	for i := 0; i < a.Len(); i++ {
		if e, d := a.Get(i).IsEquivalent(b.Get(i)); !e {
			return e, d
		}
	}
	return true, true
}
