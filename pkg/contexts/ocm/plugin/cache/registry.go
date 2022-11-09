// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/internal"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils/registry"
)

type ConstraintRegistry[T any, K registry.Key[K]] struct {
	mapping *registry.Registry[*T, K]
	elems   map[string]*registry.Registry[*T, K]
}

func (r *ConstraintRegistry[T, K]) Lookup(key K) []*T {
	return r.mapping.GetHandler(key)
}

func (r *ConstraintRegistry[T, K]) LookupFor(name string, key K) []*T {
	if name == "" {
		return r.Lookup(key)
	}
	m := r.elems[name]
	if m == nil {
		return nil
	}
	return m.GetHandler(key)
}

func NewConstraintRegistry[T internal.Element[K], K registry.Key[K]](list []T) *ConstraintRegistry[T, K] {
	reg := registry.NewRegistry[*T, K]()
	m := map[string]*registry.Registry[*T, K]{}

	for i := range list {
		d := list[i]
		nested := registry.NewRegistry[*T, K]()
		for _, c := range d.GetConstraints() {
			if c.IsValid() {
				reg.Register(c, &d)
				nested.Register(c, &d)
			}
		}
		m[d.GetName()] = nested
	}
	return &ConstraintRegistry[T, K]{reg, m}
}
