package cache

import (
	"github.com/mandelsoft/goutils/set"

	"ocm.software/ocm/api/ocm/ocmutils/registry"
	"ocm.software/ocm/api/ocm/plugin/descriptor"
)

type ConstraintRegistry[T any, K registry.Key[K]] struct {
	mapping *registry.Registry[*T, K]
	elems   map[string]*registry.Registry[*T, K]
}

func (r *ConstraintRegistry[T, K]) Lookup(key K) []*T {
	return r.mapping.LookupHandler(key)
}

func (r *ConstraintRegistry[T, K]) LookupKeys(key K) set.Set[K] {
	return r.mapping.LookupKeys(key)
}

func (r *ConstraintRegistry[T, K]) LookupFor(name string, key K) []*T {
	if name == "" {
		return r.Lookup(key)
	}
	m := r.elems[name]
	if m == nil {
		return nil
	}
	return m.LookupHandler(key)
}

func (r *ConstraintRegistry[T, K]) LookupKeysFor(name string, key K) set.Set[K] {
	if name == "" {
		return r.LookupKeys(key)
	}
	m := r.elems[name]
	if m == nil {
		return nil
	}
	return m.LookupKeys(key)
}

func NewConstraintRegistry[T descriptor.Element[K], K registry.Key[K]](list []T) *ConstraintRegistry[T, K] {
	reg := registry.NewRegistry[*T, K]()
	m := map[string]*registry.Registry[*T, K]{}

	for i := range list {
		d := list[i]
		nested := registry.NewRegistry[*T, K]()
		if len(d.GetConstraints()) == 0 {
			var zero K
			nested.Register(zero, &d)
		} else {
			for _, c := range d.GetConstraints() {
				if c.IsValid() {
					reg.Register(c, &d)
					nested.Register(c, &d)
				}
			}
		}
		m[d.GetName()] = nested
	}
	return &ConstraintRegistry[T, K]{reg, m}
}
