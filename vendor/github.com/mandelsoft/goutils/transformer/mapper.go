package transformer

type Transformer[I, O any] func(in I) O

func KeyToValue[M ~map[K]V, K comparable, V any](m M) Transformer[K, V] {
	return func(k K) V {
		return m[k]
	}
}

func IndexToValue[S ~[]V, V any](s S) Transformer[int, V] {
	return func(i int) V {
		return s[i]
	}
}
