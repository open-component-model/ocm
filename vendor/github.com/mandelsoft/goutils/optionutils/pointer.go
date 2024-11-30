package optionutils

import (
	"github.com/mandelsoft/goutils/general"
)

func PointerTo[T any](v T) *T {
	temp := v
	return &temp
}

func AsValue[T any](p *T) T {
	var r T
	if p != nil {
		r = *p
	}
	return r
}

func BoolP[T ~bool](b T) *bool {
	v := bool(b)
	return &v
}

func AsBool(b *bool, def ...bool) bool {
	if b == nil && len(def) > 0 {
		return general.Optional(def...)
	}
	return b != nil && *b
}

func ApplyOption[T any](opt *T, tgt **T) {
	if opt != nil {
		*tgt = opt
	}
}

func ApplyOptionByFunc[T any](opt *T, set func(T)) {
	if opt != nil {
		set(*opt)
	}
}

// GetOptionFlag returns the flag value used to set a bool option
// based on optionally specified explicit value(s).
// The default value is to enable the option (true).
func GetOptionFlag(list ...bool) bool {
	return general.OptionalDefaultedBool(true, list...)
}
