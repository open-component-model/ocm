package general

import (
	"fmt"
)

// Must expect a result to be provided without error.
func Must[T any](o T, err error) T {
	if err != nil {
		panic(fmt.Errorf("expected a %T, but got error %w", o, err))
	}
	return o
}

func ErrorFrom[T any](t T, err error) error {
	return err
}

func ErrorFrom2[T, U any](t T, u U, err error) error {
	return err
}

func ErrorFrom3[T, U, V any](t T, u U, v V, err error) error {
	return err
}
