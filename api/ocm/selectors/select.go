package selectors

import (
	"github.com/mandelsoft/goutils/errors"
	"github.com/mandelsoft/goutils/generics"

	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/accessors"
)

// ErrorProvider is an optional interface a Selector can offer
// to propagate an error determined setting up the selector.
// Such an error cannot be returned directly by the function
// creating the selector, because this would prohibit to
// compose selector sets as variadic arguments.
type ErrorProvider interface {
	GetError() error
}

type ErrorSelectorBase struct {
	err error
}

func NewErrorSelectorBase(err error) ErrorSelectorBase {
	return ErrorSelectorBase{err}
}

func (s *ErrorSelectorBase) GetError() error {
	return s.err
}

////////////////////////////////////////////////////////////////////////////////

type ResourceSelector interface {
	MatchResource(accessors.ElementListAccessor, accessors.ResourceAccessor) bool
}

type ResourceSelectorFunc func(accessors.ElementListAccessor, accessors.ResourceAccessor) bool

func (f ResourceSelectorFunc) MatchResource(list accessors.ElementListAccessor, a accessors.ResourceAccessor) bool {
	return f(list, a)
}

type ResourceErrorSelectorImpl struct {
	ErrorSelectorBase
	ResourceSelector
}

func NewResourceErrorSelectorImpl(s ResourceSelector, err error) *ResourceErrorSelectorImpl {
	return &ResourceErrorSelectorImpl{NewErrorSelectorBase(err), s}
}

////////////////////////////////////////////////////////////////////////////////

type SourceSelector interface {
	MatchSource(accessors.ElementListAccessor, accessors.SourceAccessor) bool
}

type SourceSelectorFunc func(accessors.ElementListAccessor, accessors.SourceAccessor) bool

func (f SourceSelectorFunc) MatchSource(list accessors.ElementListAccessor, a accessors.SourceAccessor) bool {
	return f(list, a)
}

type SourceErrorSelectorImpl struct {
	ErrorSelectorBase
	SourceSelector
}

func NewSourceErrorSelectorImpl(s SourceSelector, err error) *SourceErrorSelectorImpl {
	return &SourceErrorSelectorImpl{NewErrorSelectorBase(err), s}
}

////////////////////////////////////////////////////////////////////////////////

type ReferenceSelector interface {
	MatchReference(accessors.ElementListAccessor, accessors.ReferenceAccessor) bool
}

type ReferenceSelectorFunc func(accessors.ElementListAccessor, accessors.ReferenceAccessor) bool

func (f ReferenceSelectorFunc) MatchReference(list accessors.ElementListAccessor, a accessors.ReferenceAccessor) bool {
	return f(list, a)
}

type ReferenceErrorSelectorImpl struct {
	ErrorSelectorBase
	ReferenceSelector
}

func NewReferenceErrorSelectorImpl(s ReferenceSelector, err error) *ReferenceErrorSelectorImpl {
	return &ReferenceErrorSelectorImpl{NewErrorSelectorBase(err), s}
}

////////////////////////////////////////////////////////////////////////////////

type LabelSelector interface {
	MatchLabel(label *metav1.Label) bool
}

type LabelSelectorFunc func(label *metav1.Label) bool

func (f LabelSelectorFunc) MatchLabel(l *metav1.Label) bool {
	return f(l)
}

////////////////////////////////////////////////////////////////////////////////

func ValidateSelectors[T any](sel ...T) error {
	list := errors.ErrListf("error in selector list")
	return validateSelectors(list, sel...)
}

func ValidateSubSelectors[T any](msg string, sel ...T) error {
	list := errors.ErrList(msg)
	return validateSelectors(list, sel...)
}

func validateSelectors[T any](list *errors.ErrorList, sel ...T) error {
	for _, s := range sel {
		if p, ok := generics.TryCast[ErrorProvider](s); ok {
			list.Add(p.GetError())
		}
	}
	return list.Result()
}

////////////////////////////////////////////////////////////////////////////////

func _select[S, E any](list accessors.ElementListAccessor, m func(S, E) bool, sel ...S) []E {
	var result []E
outer:
	for i := 0; i < list.Len(); i++ {
		e := list.Get(i).(E)
		for _, s := range sel {
			if !m(s, e) {
				continue outer
			}
		}
		result = append(result, e)
	}
	return result
}

// SelectResources select resources by a set of selectors.
// It requires to be called with a compdesc.Resources  list.
func SelectResources(list accessors.ElementListAccessor, sel ...ResourceSelector) []accessors.ResourceAccessor {
	return _select(list,
		func(s ResourceSelector, e accessors.ResourceAccessor) bool {
			return s.MatchResource(list, e)
		},
		sel...)
}

// SelectSources select sources by a set of selectors.
// It requires to be called with a compdesc.Sources list.
func SelectSources(list accessors.ElementListAccessor, sel ...SourceSelector) []accessors.SourceAccessor {
	return _select(list,
		func(s SourceSelector, e accessors.SourceAccessor) bool {
			return s.MatchSource(list, e)
		}, sel...)
}

// SelectReferences select resources by a set of selectors.
// It requires to be called with a compdesc.References list.
func SelectReferences(list accessors.ElementListAccessor, sel ...ReferenceSelector) []accessors.ReferenceAccessor {
	return _select(list,
		func(s ReferenceSelector, e accessors.ReferenceAccessor) bool {
			return s.MatchReference(list, e)
		}, sel...)
}
