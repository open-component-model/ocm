package flagsetscheme

import (
	"github.com/mandelsoft/goutils/optionutils"

	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/runtime/descriptivetype"
)

// OptionTargetWrapper is used as target for option functions, it provides
// setters for fields, which should not be modifiable for a final type object.
type OptionTargetWrapper[T any] struct {
	target T
	info   *typeInfoImpl
}

func NewOptionTargetWrapper[T any](target T, info *typeInfoImpl) *OptionTargetWrapper[T] {
	return &OptionTargetWrapper[T]{
		target: target,
		info:   info,
	}
}

func (t OptionTargetWrapper[E]) SetConfigHandler(value flagsets.ConfigOptionTypeSetHandler) {
	t.info.handler = value
}

////////////////////////////////////////////////////////////////////////////////

type OptionTarget interface {
	SetConfigHandler(flagsets.ConfigOptionTypeSetHandler)
}

type TypeOption = optionutils.Option[OptionTarget]

////////////////////////////////////////////////////////////////////////////////
// options derived from descriptivetype.

func WithFormatSpec(value string) TypeOption {
	return optionutils.MapOptionTarget[OptionTarget](descriptivetype.WithFormatSpec(value))
}

func WithDescription(value string) TypeOption {
	return optionutils.MapOptionTarget[OptionTarget](descriptivetype.WithDescription(value))
}

////////////////////////////////////////////////////////////////////////////////
// additional options.

type configOption struct {
	value flagsets.ConfigOptionTypeSetHandler
}

func WithConfigHandler(value flagsets.ConfigOptionTypeSetHandler) TypeOption {
	return configOption{value}
}

func (o configOption) ApplyTo(t OptionTarget) {
	t.SetConfigHandler(o.value)
}
