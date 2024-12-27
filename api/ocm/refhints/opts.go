package refhints

import (
	"slices"

	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"
)

////////////////////////////////////////////////////////////////////////////////
// Matchers

func MatchType(typs ...string) matcher.Matcher[ReferenceHint] {
	return func(h ReferenceHint) bool {
		return slices.Contains(typs, h.GetType())
	}
}

func Equal(o ReferenceHint) matcher.Matcher[ReferenceHint] {
	d := o.Serialize()
	return func(h ReferenceHint) bool {
		return h.Serialize() == d
	}
}

func IsImplicit(h ReferenceHint) bool {
	if h == nil {
		return false
	}
	return h.GetProperty(HINT_IMPLICIT) == IMPLICIT_TRUE
}

func IsExplicit(h ReferenceHint) bool {
	if h == nil {
		return false
	}
	return !IsImplicit(h)
}

////////////////////////////////////////////////////////////////////////////////
// Filter

func Filter(hints []ReferenceHint, cond matcher.Matcher[ReferenceHint]) ReferenceHints {
	if len(hints) == 0 {
		return nil
	}
	return sliceutils.Filter(hints, cond)
}

func FilterImplicit(hints []ReferenceHint) ReferenceHints {
	return Filter(hints, IsImplicit)
}

////////////////////////////////////////////////////////////////////////////////
// Transformers

func AsImplicit[S ~[]T, T ReferenceHint](hints S) DefaultReferenceHints {
	var result DefaultReferenceHints

	for _, h := range hints {
		if IsImplicit(h) {
			result.Add(h)
		} else {
			result.Add(h.AsDefault().SetProperty(HINT_IMPLICIT, IMPLICIT_TRUE))
		}
	}
	return result
}

// AsDefault transforms a generic hint into a default hint.
// It can be used by sliceutils.Transform.
func AsDefault(h ReferenceHint) DefaultReferenceHint {
	return h.AsDefault()
}
