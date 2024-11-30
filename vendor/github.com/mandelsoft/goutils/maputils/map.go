package maputils

import (
	"cmp"
	"slices"

	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/mandelsoft/goutils/transformer"
)

type CompareFunc[E any] func(a, b E) int

// Keys provides a list of keys optionally sorted
// by a CompareFunc.
func Keys[M ~map[K]V, K comparable, V any](m M, cmp ...CompareFunc[K]) []K {
	r := []K{}

	for k := range m {
		r = append(r, k)
	}
	if len(cmp) > 0 {
		slices.SortFunc(r, cmp[0])
	}
	return r
}

// Values returns values optionally ordered by keys.
func Values[M ~map[K]V, K comparable, V any](m M, cmp ...CompareFunc[K]) []V {
	return sliceutils.Transform(Keys(m, cmp...), func(k K) V {
		return m[k]
	})
}

// OrderedKeys provides an ordered key list for maps with an ordered key type.
func OrderedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := Keys(m)
	slices.Sort(r)
	return r
}

// OrderedValues returns values optionally ordered by ordered keys.
func OrderedValues[M ~map[K]V, K cmp.Ordered, V any](m M) []V {
	return sliceutils.Transform(OrderedKeys(m), transformer.KeyToValue(m))
}

func FilterByKey[M ~map[K]V, K comparable, V any](m M, matcher matcher.Matcher[K]) M {
	if m == nil {
		return nil
	}
	r := M{}
	for k, v := range m {
		if matcher(k) {
			r[k] = v
		}
	}
	return r
}

func FilterByValue[M ~map[K]V, K comparable, V any](m M, matcher matcher.Matcher[V]) M {
	if m == nil {
		return nil
	}
	r := M{}
	for k, v := range m {
		if matcher(v) {
			r[k] = v
		}
	}
	return r
}

func FilterValues[M ~map[K]V, K comparable, V any](m M, matcher matcher.Matcher[V]) []V {
	var r []V
	for _, v := range m {
		if matcher(v) {
			r = append(r, v)
		}
	}
	return r
}

func FilterKeys[M ~map[K]V, K comparable, V any](m M, matcher matcher.Matcher[K]) []V {
	var r []V
	for k, v := range m {
		if matcher(k) {
			r = append(r, v)
		}
	}
	return r
}

type Transformer[K, V, TK, TV any] func(K, V) (TK, TV)

func KeyValueTransformer[K, V, TK, TV any](tk transformer.Transformer[K, TK], tv transformer.Transformer[V, TV]) Transformer[K, V, TK, TV] {
	return func(k K, v V) (TK, TV) {
		return tk(k), tv(v)
	}
}

func Transform[M ~map[K]V, K comparable, V any, TK comparable, TV any](in M, m Transformer[K, V, TK, TV]) map[TK]TV {
	r := map[TK]TV{}
	for k, v := range in {
		tk, tv := m(k, v)
		r[tk] = tv
	}
	return r
}

func TransformKeys[M ~map[K]V, K comparable, V any, TK comparable](in M, m transformer.Transformer[K, TK]) map[TK]V {
	r := map[TK]V{}
	for k, v := range in {
		tk := m(k)
		r[tk] = v
	}
	return r
}

func TransformedKeys[M ~map[K]V, K comparable, V any, TK comparable](in M, m transformer.Transformer[K, TK], cmp ...CompareFunc[TK]) []TK {
	r := make([]TK, len(in))
	i := 0
	for k := range in {
		tk := m(k)
		r[i] = tk
		i++
	}
	if len(cmp) > 0 {
		slices.SortFunc(r, cmp[0])
	}
	return r
}

func OrderedTransformedKeys[M ~map[K]V, K comparable, V any, TK cmp.Ordered](in M, m transformer.Transformer[K, TK]) []TK {
	r := make([]TK, len(in))
	i := 0
	for k := range in {
		tk := m(k)
		r[i] = tk
		i++
	}
	slices.Sort(r)
	return r
}

func TransformValues[M ~map[K]V, K comparable, V any, TV any](in M, m transformer.Transformer[V, TV]) map[K]TV {
	r := map[K]TV{}
	for k, v := range in {
		tv := m(v)
		r[k] = tv
	}
	return r
}

func TransformedValues[M ~map[K]V, K comparable, V any, TV any](in M, m transformer.Transformer[V, TV], cmp ...CompareFunc[TV]) []TV {
	r := make([]TV, len(in))
	i := 0
	for _, v := range in {
		tv := m(v)
		r[i] = tv
		i++
	}
	if len(cmp) > 0 {
		slices.SortFunc(r, cmp[0])
	}
	return r
}

func OrderedTransformedValues[M ~map[K]V, K comparable, V any, TV cmp.Ordered](in M, m transformer.Transformer[V, TV]) []TV {
	r := make([]TV, len(in))
	i := 0
	for _, v := range in {
		tv := m(v)
		r[i] = tv
		i++
	}
	slices.Sort(r)
	return r
}
