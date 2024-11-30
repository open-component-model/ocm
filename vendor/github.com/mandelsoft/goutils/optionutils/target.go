package optionutils

// ///////////////////////////////////////////////////////////////////////////(//
// if the option target is an interface, it is easily possible to
// provide new targets with more options just by extending the
// target interface. The option consumer the accepts options for
// the target interface.
// To be able to reuse options from the base target interface
// a wrapper option implementation is required which implements
// the extended option interface and maps it to the base option
// interface.
//
// The following mechanism requires option targets W and B to be
// interface types to get applied.
//
// If W does not implement B the option is ignored, when
// applying to the new option target W.
// This can be used together with FilterMappedOptions
// to pass options formally mapped to a new option target
// to a consumer of the original option type, instead of
// applying it as part to the new option target.
type targetInterfaceWrapper[B any, W any /*B*/] struct {
	option Option[B]
}

var _ MappedOption[int] = (*targetInterfaceWrapper[int, string])(nil)

func (w *targetInterfaceWrapper[B, W]) ApplyTo(opts W) {
	var i any = opts
	if base, ok := i.(B); ok {
		w.option.ApplyTo(base)
	}
}

func (w *targetInterfaceWrapper[B, W]) Unwrap() Option[B] {
	return w.option
}

// MappedOption is the interface for an option mapped to another
// option target. It returns the original option.
type MappedOption[B any] interface {
	Unwrap() Option[B]
}

// MapOptionTarget maps the option target interface from
// B to W, hereby, W must be a subtype of B, which cannot be
// expressed with Go generics (Type constraint should be W B).
// If this constraint is not met, there will be a runtime error.
func MapOptionTarget[W, B any](opt Option[B]) Option[W] {
	return &targetInterfaceWrapper[B, W]{
		opt,
	}
}

// FilterMappedOptions filters options (for S) for options mapped to
// another option interface given by type parameter T.
func FilterMappedOptions[T any, S any](opts ...S) []Option[T] {
	var result []Option[T]
	for _, o := range opts {
		var i any = o
		if m, ok := i.(MappedOption[T]); ok {
			result = append(result, m.Unwrap())
		}
	}
	return result
}
