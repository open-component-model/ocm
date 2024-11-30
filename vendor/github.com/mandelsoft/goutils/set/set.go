// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package set

import (
	"github.com/mandelsoft/goutils/maputils"
)

type Set[K comparable] map[K]struct{}

func New[K comparable](keys ...K) Set[K] {
	return Set[K]{}.Add(keys...)
}

func (s Set[K]) Add(keys ...K) Set[K] {
	for _, k := range keys {
		s[k] = struct{}{}
	}
	return s
}

func (s Set[K]) AddAll(set Set[K]) Set[K] {
	for k := range set {
		s[k] = struct{}{}
	}
	return s
}

func (s Set[K]) Delete(keys ...K) Set[K] {
	for _, k := range keys {
		delete(s, k)
	}
	return s
}

func (s Set[K]) DeleteAll(set Set[K]) Set[K] {
	for k := range set {
		delete(s, k)
	}
	return s
}

func (s Set[K]) Contains(keys ...K) bool {
	for _, k := range keys {
		if _, ok := s[k]; !ok {
			return false
		}
	}
	return true
}

func (s Set[K]) AsArray() []K {
	keys := []K{}
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s Set[K]) Has(key K) bool {
	_, ok := s[key]
	return ok
}

func Keys[K comparable, V any](m map[K]V, cmp maputils.CompareFunc[K]) []K {
	return maputils.Keys(m, cmp)
}

func KeySet[K comparable, V any](m map[K]V) Set[K] {
	s := Set[K]{}
	for k := range m {
		s.Add(k)
	}
	return s
}
