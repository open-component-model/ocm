package utils

import (
	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/ioutils"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/optionutils"
	"github.com/mandelsoft/goutils/stringutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/modern-go/reflect2"
)

// Deprecated: use maputils.OrderedKeys.
func StringMapKeys[K ~string, E any](m map[K]E) []K {
	return maputils.OrderedKeys(m)
}

// Optional returns the first optional non-zero element given as variadic argument,
// if given, or the zero element as default.
// Deprecated: use general.Optional or general.OptionalNonZero.
func Optional[T any](list ...T) T {
	return general.Optional(list...)
}

// OptionalDefaulted returns the first optional non-nil element given as variadic
// argument, or the given default element. For value types a given zero
// argument is excepted, also.
// Deprecated: use general.OptionalNonZeroDefaulted or general.OptionaDefaulted.
func OptionalDefaulted[T any](def T, list ...T) T {
	return general.OptionalDefaulted(def, list...)
}

// OptionalDefaultedBool checks all args for true. If arg is given
// the given default is returned.
// Deprecated: use general.OptionalDefaultedBool.
func OptionalDefaultedBool(def bool, list ...bool) bool {
	return general.OptionalDefaultedBool(def, list...)
}

// Deprecated: use optionutils.BoolP.
func BoolP[T ~bool](b T) *bool {
	return optionutils.BoolP(b)
}

// Deprecated: use optionutils.AsBool.
func AsBool(b *bool, def ...bool) bool {
	if b == nil && len(def) > 0 {
		return Optional(def...)
	}
	return b != nil && *b
}

// GetOptionFlag returns the flag value used to set a bool option
// based on optionally specified explicit value(s).
// The default value is to enable the option (true).
// Deprecated: use optionutils.GetOptionFlag.
func GetOptionFlag(list ...bool) bool {
	return optionutils.GetOptionFlag(list...)
}

// Deprecated: use reflect2.IsNil.
func IsNil(o interface{}) bool {
	return reflect2.IsNil(o)
}

// Must expect a result to be provided without error.
// Deprecated: use general.Must
func Must[T any](o T, err error) T {
	return general.Must[T](o, err)
}

// Deprecated: use stringutils.IndentLines.
func IndentLines(orig string, gap string, skipfirst ...bool) string {
	return stringutils.IndentLines(orig, gap, skipfirst...)
}

// Deprecated: use stringutils.JoinIndentLines.
func JoinIndentLines(orig []string, gap string, skipfirst ...bool) string {
	return stringutils.JoinIndentLines(orig, gap, skipfirst...)
}

// ResolvePath handles the ~ notation for the home directory.
// Deprecated: use ioutils.ResolvePath.
func ResolvePath(path string) (string, error) {
	return ioutils.ResolvePath(path)
}

// Deprecated: use optionutils.ResolveData.
func ResolveData(in string, fss ...vfs.FileSystem) ([]byte, error) {
	return optionutils.ResolveData(in, fss...)
}

// Deprecated: use optionutils.ReadFile.
func ReadFile(in string, fss ...vfs.FileSystem) ([]byte, error) {
	return optionutils.ReadFile(in, fss...)
}
