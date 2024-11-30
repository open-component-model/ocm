package optionutils

type Option[T any] interface {
	ApplyTo(T)
}

// NoOption is an option that does nothing. This can be returned by Option Factory functions instead of nil to avoid
// consecutive nil pointer exceptions.
type NoOption[T any] struct{}

func (NoOption[T]) ApplyTo(T) {}

// EvalOptions applies options to a new options object
// and returns this object.
// O must be a struct type.
func EvalOptions[O any](opts ...Option[*O]) *O {
	var eff O
	ApplyOptions(&eff, opts...)
	return &eff
}

// ApplyOptions applies options to
// an option target O. O must either
// be a target interface type or a target struct
// pointer type.
func ApplyOptions[O any](opts O, list ...Option[O]) {
	for _, opt := range list {
		if opt != nil {
			opt.ApplyTo(opts)
		}
	}
}
