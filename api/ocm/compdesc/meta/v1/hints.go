package v1

import (
	"maps"
	"slices"
	"strings"

	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	HINT_TYPE      = "type"
	HINT_REFERENCE = "reference"
)

func MatchHintType(typs ...string) matcher.Matcher[ReferenceHint] {
	return func(h ReferenceHint) bool {
		return slices.Contains(typs, h.GetType())
	}
}

// ReferenceHintProvider gets the optional hints
// for list of possible hint types.
type ReferenceHintProvider interface {
	GetReferenceHints(typs ...string) []ReferenceHint
}

type ReferenceHints []ReferenceHint

var _ ReferenceHintProvider = (ReferenceHints)(nil)

func (h ReferenceHints) Copy() ReferenceHints {
	var result ReferenceHints

	for _, v := range h {
		result = append(result, v.Copy())
	}
	return result
}

func (h ReferenceHints) GetReferenceHints(typs ...string) []ReferenceHint {
	if len(typs) == 0 {
		return slices.Clone(h)
	}
	return sliceutils.Filter(h, MatchHintType(typs...))
}

func (h ReferenceHints) GetCompatReferenceHint() string {
	return GetCompatReferenceHint(h)
}

func GetCompatReferenceHint[H ReferenceHint](h []H) string {
	for _, e := range h {
		return e.GetCompatReferenceHint()
	}
	return ""
}

type ReferenceHint interface {
	runtime.TypedObject

	Copy() ReferenceHint
	GetCompatReferenceHint() string
	GetReference() string
	String() string

	GetProperty(name string) string

	AsDefault() DefaultReferenceHint
}

type DefaultReferenceHint map[string]string

var _ ReferenceHint = DefaultReferenceHint{}

func NewReferenceHint(typ, ref string) DefaultReferenceHint {
	return DefaultReferenceHint{
		HINT_TYPE:      typ,
		HINT_REFERENCE: ref,
	}
}

func (h DefaultReferenceHint) GetType() string {
	return h[HINT_TYPE]
}

func (h DefaultReferenceHint) Copy() ReferenceHint {
	return maps.Clone(h)
}

func (h DefaultReferenceHint) AsDefault() DefaultReferenceHint {
	return h
}

func (h DefaultReferenceHint) GetProperty(name string) string {
	return h[name]
}

func (h DefaultReferenceHint) GetCompatReferenceHint() string {
	if p, ok := h[HINT_REFERENCE]; ok {
		if t, ok := h[HINT_TYPE]; ok {
			return t + "::" + p
		} else {
			return p
		}
	}
	return ""
}

func (h DefaultReferenceHint) GetReference() string {
	if p, ok := h[HINT_REFERENCE]; ok {
		return p
	}

	sep := ""
	s := ""
	for _, k := range maputils.OrderedKeys(h) {
		if k != HINT_TYPE {
			s = s + sep + k + "=" + h[k]
			sep = ","
		}
	}
	return s
}

func (h DefaultReferenceHint) String() string {
	sep := ""
	s := ""
	t := h[HINT_TYPE]
	if t != "" {
		s = t + "::"
	}
	for _, k := range maputils.OrderedKeys(h) {
		if k != HINT_TYPE {
			s = s + sep + k + "=" + h[k]
			sep = ","
		}
	}
	return s
}

func StringToHint(s string) ReferenceHint {
	i := strings.Index(s, "::")
	if i >= 0 {
		return NewReferenceHint(s[:i], s[i+2:])
	}
	return DefaultReferenceHint{
		HINT_REFERENCE: s,
	}
}
