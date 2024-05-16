package optionutils

import (
	"github.com/open-component-model/ocm/pkg/utils"
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

func AsBool(b *bool, def ...bool) bool {
	return utils.AsBool(b, def...)
}

func ApplyOption[T any](opt *T, tgt **T) {
	if opt != nil {
		*tgt = opt
	}
}

// GetOptionFlag returns the flag value used to set a bool option
// based on optionally specified explicit value(s).
// The default value is to enable the option (true).
func GetOptionFlag(list ...bool) bool {
	return utils.OptionalDefaultedBool(true, list...)
}
