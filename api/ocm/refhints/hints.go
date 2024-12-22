package refhints

import (
	"maps"
	"slices"
	"strings"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/matcher"
	"github.com/mandelsoft/goutils/sliceutils"

	"ocm.software/ocm/api/utils/runtime"
)

const (
	// HINT_TYPE describes the type of the hinf.
	// For example oci or maven.
	HINT_TYPE = "type"
	// HINT_REFERENCE is the default field holding
	// a reference format according to the hint type.
	HINT_REFERENCE = "reference"
	// HINT_IMPLICIT is a flag field indicating
	// that the hint is implicitly provided by an access
	// method. Explicit hints are hints consciously
	// provided as part of the artifact metadata.
	// This attribute should not be serialized for
	// persisting hints, only for passing them as string. It
	// is provided by the access method, only.
	HINT_IMPLICIT = "implicit"

	IMPLICIT_TRUE = "true"
)

func MatchType(typs ...string) matcher.Matcher[ReferenceHint] {
	return func(h ReferenceHint) bool {
		return slices.Contains(typs, h.GetType())
	}
}

// ReferenceHints is list of hints.
// Notaion: a sequence of hint notations separated by a ;.
type ReferenceHints []ReferenceHint

func NewHints(f func(ref string, implicit ...bool) ReferenceHint, ref string, implicit ...bool) ReferenceHints {
	return ReferenceHints{f(ref, implicit...)}
}

func (h *ReferenceHints) Add(hints ...ReferenceHint) {
	*h = sliceutils.AppendUniqueFunc(*h, runtime.MatchType[ReferenceHint], hints...)
}

func (h ReferenceHints) Copy() ReferenceHints {
	var result ReferenceHints

	for _, v := range h {
		result = append(result, v.Copy())
	}
	return result
}

func (h ReferenceHints) GetReferenceHints(typs ...string) ReferenceHints {
	if len(typs) == 0 {
		return h.Copy()
	}
	return sliceutils.Filter(h, MatchType(typs...))
}

// GetReference returns the hint for the first available hint type
// of the given type list.
func (h ReferenceHints) GetReferenceHint(typs ...string) ReferenceHint {
	if len(typs) == 0 {
		return nil
	}
	hints := sliceutils.Filter(h, MatchType(typs...))
	if len(hints) == 0 {
		return nil
	}
	return hints[0]
}

// Serialize provides a string representation. The implicit
// attribute is only serialized, if it is called with true.
func (h ReferenceHints) Serialize(implicit ...bool) string {
	return HintsToString(h, implicit...)
}

func HintsToString(hints []ReferenceHint, implicit ...bool) string {
	sep := ""
	r := ""
	for _, h := range hints {
		r = r + sep + h.Serialize(general.Optional(implicit...))
		sep = ";"
	}
	return r
}

// ReferenceHint is the internal representation of
// a reference hint used to recreate a type specific
// identity for artifacts given as blob, which are
// uploaded to a technology specif registry, again.
// It consists of a set of simple properties, whose names
// must consist of letters or digits, only.
// Notation:
//   - <typ>::<value>
//   - <value>
//   - <typ>::<attr>=<value>{,<attr>=<value>}"
//
// If the value contains any of ",;\ those characters must
// be escaped with \ and the value must be in double quotes.
// Otherwise, the double quotes are optional.
type ReferenceHint interface {
	runtime.TypedObject

	GetReference() string
	// GetReefernce provides the default reference attribute.
	Copy() ReferenceHint

	// Serialize provides a string representation. The implicit
	// attribute is only serialized, if it is called with true.
	Serialize(implicit ...bool) string

	GetProperty(name string) string

	AsDefault() DefaultReferenceHint
}

func IsImplicitHint(h ReferenceHint) bool {
	if h == nil {
		return false
	}
	return h.GetProperty(HINT_IMPLICIT) == IMPLICIT_TRUE
}

func FilterImplicit(hints []ReferenceHint) ReferenceHints {
	if len(hints) == 0 {
		return nil
	}
	return sliceutils.Filter(hints, IsImplicitHint)
}

// GetReference returns the default reference hint attribute
// for the first given type available in the hint list.
func GetReference(hints []ReferenceHint, typs ...string) string {
	if len(hints) == 0 {
		return ""
	}
	h := ReferenceHints(hints).GetReferenceHint(typs...)
	if h == nil {
		return ""
	}
	return h.GetReference()
}

type DefaultReferenceHint map[string]string

var _ ReferenceHint = DefaultReferenceHint{}

func New(typ, ref string, implicit ...bool) DefaultReferenceHint {
	h := DefaultReferenceHint{
		HINT_REFERENCE: ref,
	}
	if typ != "" {
		h[HINT_TYPE] = typ
	}
	if general.Optional(implicit...) {
		h[HINT_IMPLICIT] = IMPLICIT_TRUE
	}
	return h
}

func DefaultHint(ref string, implicit ...bool) ReferenceHint {
	return New("", ref, implicit...)
}

func (h DefaultReferenceHint) GetType() string {
	if h == nil {
		return ""
	}
	return h[HINT_TYPE]
}

func (h DefaultReferenceHint) Copy() ReferenceHint {
	return maps.Clone(h)
}

func (h DefaultReferenceHint) AsDefault() DefaultReferenceHint {
	return h
}

func (h DefaultReferenceHint) GetProperty(name string) string {
	if h == nil {
		return ""
	}
	return h[name]
}

func (h DefaultReferenceHint) SetProperty(name, val string) DefaultReferenceHint {
	if h != nil {
		h[name] = val
	}
	return h
}

func (h DefaultReferenceHint) GetReference() string {
	if h == nil {
		return ""
	}
	return h[HINT_REFERENCE]
}

func (h DefaultReferenceHint) Serialize(implicit ...bool) string {
	if h == nil {
		return ""
	}
	if !general.Optional(implicit...) {
		if h.GetProperty(HINT_IMPLICIT) != "" {
			h = maps.Clone(h)
			delete(h, HINT_IMPLICIT)
		}
	}
	sep := ""
	s := ""
	t, typefound := h[HINT_TYPE]
	if t != "" {
		s = t + "::"

		if r, ok := h[HINT_REFERENCE]; ok && len(h) == 2 {
			return s + escapeHintValue(r)
		}
	} else {
		if r, ok := h[HINT_REFERENCE]; ok && ((!typefound && len(h) == 1) || (typefound && len(h) == 2)) {
			return escapeHintValue(r)
		}
	}
	for _, k := range maputils.OrderedKeys(h) {
		if k != HINT_TYPE {
			s = s + sep + k + "=" + escapeHintValue(h[k])
			sep = ","
		}
	}
	return s
}

func escapeHintValue(v string) string {
	if !strings.ContainsAny(v, "\",;") {
		return v
	}

	r := "\""
	for _, c := range v {
		if c == '\\' || c == '"' {
			r += "\\"
		}
		r += string(c)
	}
	return r + "\""
}

func newHint(impl bool) DefaultReferenceHint {
	h := DefaultReferenceHint{}
	if impl {
		h[HINT_IMPLICIT] = IMPLICIT_TRUE
	}
	return h
}

// ParseHints parses a string containing servialized reference hints,
// If implicit is set to true, the implicit attribute is set
func ParseHints(v string, implicit ...bool) ReferenceHints {
	var hints ReferenceHints

	var prop string
	var val string

	var hint DefaultReferenceHint
	state := -1
	start := 0
	mask := false
	impl := general.Optional(implicit...)
	for i, c := range v {
		switch state {
		case -1:
			if c == '"' {
				hint = newHint(impl)
				prop = HINT_REFERENCE
				start = i + 1
				state = 5
			} else {
				state = 0
			}
			fallthrough
		case 0: // type
			if c == ':' {
				state = 1
			}
		case 1: // colon
			if c == ':' {
				hint = newHint(impl).SetProperty(HINT_TYPE, v[start:i-1])
				start = i + 1
				state = 7
			} else {
				state = 0
			}
		case 7: // prop start
			if c == '"' {
				val = ""
				prop = HINT_REFERENCE
				state = 5
				start = i + 1
				continue
			}
			state = 2
			fallthrough
		case 2: // prop
			switch c {
			case '=':
				prop = v[start:i]
				start = i + 1
				state = 3
			case ';':
				hint[HINT_REFERENCE] = v[start:i]
				hints = append(hints, hint)
				hint = nil
				state = -1
				start = i + 1
			}
		case 3: // value start
			if c == '"' {
				val = ""
				state = 5
				start = i + 1
			} else {
				state = 4
				start = i
			}
		case 4: // plain value
			if c == ',' || c == ';' {
				hint[prop] = v[start:i]
				start = i + 1
				if c == ';' {
					hints = append(hints, hint)
					hint = nil
					state = -1
				} else {
					state = 2
				}
			}
		case 5: // escaped value
			if mask {
				mask = false
			} else {
				if c == '\\' {
					mask = true
					continue
				}
				if c == '"' {
					hint[prop] = val
					state = 6
				}
			}
			val += string(c)
		case 6: // end escaped
			if c == ',' {
				start = i + 1
				state = 2
			}
			if c == ';' {
				hints = append(hints, hint)
				hint = nil
				start = i + 1
				state = -1
			}
		}
	}

	switch state {
	case 0, 1:
		hint = newHint(impl).SetProperty(HINT_REFERENCE, v[start:])
	case 2:
		hint[HINT_REFERENCE] = v[start:]
	case 3:
		hint[prop] = ""
	case 4:
		hint[prop] = v[start:]
	case 5:
		hint[prop] = v[start:]
	case 6:
	}
	hints = append(hints, hint)
	return hints
}

func Join(hints ...[]ReferenceHint) ReferenceHints {
	var result []ReferenceHint
	for _, h := range hints {
		result = sliceutils.AppendUniqueFunc(result, runtime.MatchType[ReferenceHint], h...)
	}
	return result
}
