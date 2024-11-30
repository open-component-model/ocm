package optionutils

import (
	"github.com/mandelsoft/goutils/reflectutils"
	"github.com/modern-go/reflect2"
)

// ArbitraryOption is just generiuc interface to
// indicate the purpose for parameters.
// because Go does not support parameter overloading,
// different option targets may need to use different apply
// method names, it an option should be usable for multiple target
// types. Therefore, a simple generic common Eval method is not possible,
// because it must call the method with a dedicated name for the intended target.
// THis package provided some generic implementation being able to
// call apply methods based on a given non-standard option interface.
// This type is used to indicate such type. The concrete type must be an
// option interface type with a single apply method.
type ArbitraryOption interface {
}

// EvalArbitraryOptions applies options to a new options object
// and returns this object.
// O must be a struct type.
func EvalArbitraryOptions[I ArbitraryOption, O any](opts ...I) *O {
	var eff O
	ApplyArbitraryOptions(&eff, opts...)
	return &eff
}

// ApplyArbitraryOptions applies options to
// an option target O. O must either
// be a target interface type or a target struct
// pointer type. I is the option interface declaring the apply
// method.
func ApplyArbitraryOptions[I ArbitraryOption, O any](opts O, list ...I) {
	m := reflectutils.GetInterfaceMethod[I]()
	for _, opt := range list {
		if !reflect2.IsNil(opt) {
			reflectutils.CallMethodByNameVA(m.Name, opt, opts)
		}
	}
}
