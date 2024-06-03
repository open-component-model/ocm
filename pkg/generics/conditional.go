package generics

func Conditional[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
