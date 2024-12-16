package v1

import (
	"maps"
	"slices"

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

type ReferenceHints []ReferenceHint

func (h ReferenceHints) Copy() ReferenceHints {
	var result ReferenceHints

	for _, v := range h {
		result = append(result, v.Copy())
	}
	return result
}

func (h ReferenceHints) Get(typs ...string) []ReferenceHint {
	if len(typs) == 0 {
		return slices.Clone(h)
	}
	return sliceutils.Filter(h, MatchHintType(typs...))
}

type ReferenceHint interface {
	runtime.TypedObject

	Copy() ReferenceHint

	GetReference() string
	String() string

	GetProperty(name string) string

	AsDefault() DefaultReferenceHint
}

type DefaultReferenceHint map[string]string

var _ ReferenceHint = DefaultReferenceHint{}

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
