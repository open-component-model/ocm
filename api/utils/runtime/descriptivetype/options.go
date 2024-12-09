package descriptivetype

import (
	"github.com/mandelsoft/goutils/optionutils"
)

////////////////////////////////////////////////////////////////////////////////

// OptionTargetWrapper is used as target for option functions, it provides
// setters for fields, which should not be modifiable for a final type object.
type OptionTargetWrapper[T any] struct {
	target T
	info   *typeInfoImpl
}

func NewOptionTargetWrapper[T any](target T, info *typeInfoImpl) *OptionTargetWrapper[T] {
	return &OptionTargetWrapper[T]{target, info}
}

func (t *OptionTargetWrapper[T]) SetDescription(value string) {
	t.info.description = value
}

func (t *OptionTargetWrapper[T]) SetFormat(value string) {
	t.info.format = value
}

func (t *OptionTargetWrapper[T]) Target() T {
	return t.target
}

////////////////////////////////////////////////////////////////////////////////
// Descriptive Type Options

type OptionTarget interface {
	SetFormat(string)
	SetDescription(string)
}

type Option = optionutils.Option[OptionTarget]

////////////////////////////////////////////////////////////////////////////////

type formatOption struct {
	value string
}

func WithFormatSpec(value string) Option {
	return formatOption{value}
}

func (o formatOption) ApplyTo(t OptionTarget) {
	t.SetFormat(o.value)
}

////////////////////////////////////////////////////////////////////////////////

type descriptionOption struct {
	value string
}

func WithDescription(value string) Option {
	return descriptionOption{value}
}

func (o descriptionOption) ApplyTo(t OptionTarget) {
	t.SetDescription(o.value)
}
