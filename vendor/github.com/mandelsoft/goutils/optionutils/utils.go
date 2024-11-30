package optionutils

import (
	"github.com/mandelsoft/goutils/sliceutils"
)

// WithDefaults prepends a given option list by an arbitrary
// number of default options.
// Those options will be evaluated before the given option set.
// They will be overridden later by the explicitly specified option set.
//
// For example:
//
//	func FuncWithOptions(ctx SomeType, opts...Option) {
//	   doSomethingWithDefaultedOptions(WithDefaults(opts, WithOther(ctx.Other())...)
//	}
func WithDefaults[O any](opts []O, defaults ...O) []O {
	return sliceutils.CopyAppend(defaults, opts...)
}
