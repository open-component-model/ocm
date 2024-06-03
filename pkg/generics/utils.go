package generics

func Pointer[T any](t T) *T {
	return &t
}

func Value[T any](t *T) T {
	if t != nil {
		return *t
	}
	var v T
	return v
}
