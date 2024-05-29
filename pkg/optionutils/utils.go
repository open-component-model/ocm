package optionutils

import "github.com/mandelsoft/goutils/sliceutils"

func WithDefaults[O any](opts []O, defaults ...O) []O {
	return sliceutils.CopyAppend(defaults, opts...)
}
