package srcsel

import (
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/selectors/accessors"
)

var (
	_ selectors.ErrorProvider = (and)(nil)
	_ selectors.ErrorProvider = (or)(nil)
	_ selectors.ErrorProvider = (*not)(nil)
)

////////////////////////////////////////////////////////////////////////////////

type and []Selector

func (a and) MatchSource(list accessors.ElementListAccessor, e accessors.SourceAccessor) bool {
	for _, o := range a {
		if !o.MatchSource(list, e) {
			return false
		}
	}
	return true
}

func (a and) GetError() error {
	return selectors.ValidateSubSelectors("and", []Selector(a)...)
}

func And(operands ...Selector) Selector {
	return and(operands)
}

////////////////////////////////////////////////////////////////////////////////

type or []Selector

func (a or) MatchSource(list accessors.ElementListAccessor, e accessors.SourceAccessor) bool {
	for _, o := range a {
		if o.MatchSource(list, e) {
			return true
		}
	}
	return false
}

func (a or) GetError() error {
	return selectors.ValidateSubSelectors("or", []Selector(a)...)
}

func Or(operands ...Selector) Selector {
	return or(operands)
}

////////////////////////////////////////////////////////////////////////////////

type not struct {
	Selector
}

func (a *not) MatchReference(list accessors.ElementListAccessor, e accessors.SourceAccessor) bool {
	return !a.Selector.MatchSource(list, e)
}

func (a *not) GetError() error {
	return selectors.ValidateSubSelectors("not", a.Selector)
}

func Not(operand Selector) Selector {
	return &not{operand}
}
