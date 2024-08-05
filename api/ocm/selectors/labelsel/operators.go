package labelsel

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors"
)

var (
	_ selectors.ErrorProvider = (or)(nil)
	_ selectors.ErrorProvider = (*not)(nil)
)

////////////////////////////////////////////////////////////////////////////////

func And(sel ...Selector) *selectors.LabelErrPropSelectorImpl {
	return selectors.Label(sel...)
}

////////////////////////////////////////////////////////////////////////////////

type or []Selector

func (a or) MatchLabel(l *v1.Label) bool {
	for _, o := range a {
		if o.MatchLabel(l) {
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
	sel Selector
}

func (a *not) MatchLabel(l *v1.Label) bool {
	return !a.sel.MatchLabel(l)
}

func (a *not) GetError() error {
	return selectors.ValidateSubSelectors("not", a.sel)
}

func Not(operand Selector) Selector {
	return &not{operand}
}
