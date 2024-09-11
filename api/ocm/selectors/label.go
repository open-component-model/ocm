package selectors

import (
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/selectors/accessors"
)

func SelectLabels(labels v1.Labels, sel ...LabelSelector) ([]v1.Label, error) {
	err := ValidateSelectors(sel...)
	if err != nil {
		return nil, err
	}
	return GetLabels(labels, sel...), nil
}

func GetLabels(labels v1.Labels, sel ...LabelSelector) v1.Labels {
	result := v1.Labels{}
outer:
	for _, l := range labels {
		for _, s := range sel {
			if !s.MatchLabel(&l) {
				continue outer
			}
		}
		result = append(result, l)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

type LabelSelectorImpl struct {
	LabelSelector
}

func (i *LabelSelectorImpl) MatchResource(list accessors.ElementListAccessor, a accessors.ResourceAccessor) bool {
	return len(GetLabels(a.GetMeta().GetLabels(), i)) > 0
}

func (i *LabelSelectorImpl) MatchSource(list accessors.ElementListAccessor, a accessors.SourceAccessor) bool {
	return len(GetLabels(a.GetMeta().GetLabels(), i)) > 0
}

func (i *LabelSelectorImpl) MatchReference(list accessors.ElementListAccessor, a accessors.ReferenceAccessor) bool {
	return len(GetLabels(a.GetMeta().GetLabels(), i)) > 0
}

type LabelErrPropSelectorImpl struct {
	LabelSelectorImpl
}

func (l *LabelErrPropSelectorImpl) GetError() error {
	if e, ok := l.LabelSelector.(ErrorProvider); ok {
		return e.GetError()
	}
	return nil
}

type LabelErrorSelectorImpl struct {
	ErrorSelectorBase
	LabelSelectorImpl
}

func NewLabelErrorSelectorImpl(s LabelSelector, err error) *LabelErrorSelectorImpl {
	return &LabelErrorSelectorImpl{NewErrorSelectorBase(err), LabelSelectorImpl{s}}
}

////////////////////////////////////////////////////////////////////////////////

type label []LabelSelector

func (s label) MatchLabel(l *v1.Label) bool {
	for _, n := range s {
		if !n.MatchLabel(l) {
			return false
		}
	}
	return true
}

func (s label) GetError() error {
	return ValidateSubSelectors("and", []LabelSelector(s)...)
}

func Label(sel ...LabelSelector) *LabelErrPropSelectorImpl {
	return &LabelErrPropSelectorImpl{LabelSelectorImpl{label(sel)}}
}
