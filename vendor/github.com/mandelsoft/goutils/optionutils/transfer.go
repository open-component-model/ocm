package optionutils

import (
	"slices"

	"github.com/mandelsoft/goutils/general"
)

// Transfer transfers an option value from aan
// option object to a target value.
// If the option value in initial is is not transferred.
func Transfer[T comparable](t *T, v T) {
	var zero T

	if v != zero {
		*t = v
	}
}

// TransferOptional transfers an optional option value
// given as pointer type (nil means not set) from an
// option object to a target value, if it is set.
func TransferOptional[T comparable](t *T, v *T) {
	if v != nil {
		*t = *v
	}
}

// TransferSlice transfers an optional slice option value
// from an option object to a target value, if it is set.
// It is assumed to be set, if it is non-nil or (if empty is set to true)
// it is an empty slice.
func TransferSlice[T comparable](t *[]T, v []T, empty ...bool) {
	if v != nil || (len(v) == 0 && general.Optional(empty...)) {
		*t = slices.Clone(v)
	}
}
