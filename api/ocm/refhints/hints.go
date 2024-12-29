package refhints

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	"github.com/mandelsoft/goutils/general"
	"github.com/mandelsoft/goutils/maputils"
	"github.com/mandelsoft/goutils/sliceutils"
	"github.com/modern-go/reflect2"

	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/runtime"
)

// NONE describes non-existing hints.
var NONE = []ReferenceHint(nil)

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

type AnyReferenceHint = interface{}

// ReferenceHints is list of hints.
// Notaion: a sequence of hint notations separated by a ;.
type ReferenceHints []ReferenceHint

// DefaultList provide a list with a single typed default reference hint.
// It uses the hint creation function provided by the hint technologies
// supporting a default hint.
func DefaultList(f func(ref string, implicit ...bool) ReferenceHint, ref string, implicit ...bool) ReferenceHints {
	return ReferenceHints{f(ref, implicit...)}
}

func (h *ReferenceHints) Add(hints ...ReferenceHint) {
	AddUnique(h, hints...)
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
	hints := Filter(h, MatchType(typs...))
	if len(hints) == 0 {
		return nil
	}
	return hints[0]
}

// Serialize provides a string representation. The implicit
// attribute is only serialized, if it is called with true.
func (h ReferenceHints) Serialize(implicit ...bool) string {
	return Serialize(h, implicit...)
}

// Serialize provides a string representation for any kind of hint slice.
// The implicit attribute is only serialized, if it is called with true.
func Serialize[S ~[]T, T ReferenceHint](hints S, implicit ...bool) string {
	if len(hints) == 0 {
		return ""
	}

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

// Serialize see ParseHints for serialization rules.
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

////////////////////////////////////////////////////////////////////////////////

type DefaultReferenceHints []DefaultReferenceHint

func (h *DefaultReferenceHints) Add(hints ...ReferenceHint) {
	AddUnique(h, sliceutils.Transform(hints, AsDefault)...)
}

func (h DefaultReferenceHints) Copy() ReferenceHints {
	var result ReferenceHints

	for _, v := range h {
		result = append(result, v.Copy())
	}
	return result
}

// Serialize provides a string representation. The implicit
// attribute is only serialized, if it is called with true.
func (h DefaultReferenceHints) Serialize(implicit ...bool) string {
	return Serialize(h, implicit...)
}

var _ json.Marshaler = DefaultReferenceHints{}

func (h DefaultReferenceHints) MarshalJSON() ([]byte, error) {
	return json.Marshal(([]DefaultReferenceHint)(h))
}

var _ json.Unmarshaler = &DefaultReferenceHints{}

// UnmarshalJSON excepts the serialized form or the list form.
func (h *DefaultReferenceHints) UnmarshalJSON(data []byte) error {
	var in []DefaultReferenceHint

	err := json.Unmarshal(data, &in)
	if err == nil {
		*h = DefaultReferenceHints(in)
		return nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*h = sliceutils.Transform(ParseHints(s), AsDefault)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

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

////////////////////////////////////////////////////////////////////////////////#

// JoinUnique joins multiple hint lists, where the first occurrence of a
// hint type takes precedence.
func JoinUnique(hints ...[]ReferenceHint) ReferenceHints {
	var result []ReferenceHint
	for _, h := range hints {
		AddUnique(&result, h...)
	}
	return result
}

// AddUnique adds hints to hint list, whode type is not yet present in the list.
func AddUnique[S ~[]T, T ReferenceHint](hints *S, add ...T) {
	*hints = sliceutils.AppendUniqueFunc(*hints, runtime.MatchType[T], add...)
}

// HintsFor provide hints for several kinds of hint specifications:
//   - string (uses ParseHint)
//   - single ReferenceHint
//   - []ReferenceHint
//   - ReferenceHints
//   - []DefaultReferenceHint
//   - DefaultReferenceHints
//   - map[string]string
//   - common.Properties
func HintsFor(spec interface{}, implicit ...bool) (ReferenceHints, error) {
	if reflect2.IsNil(spec) {
		return nil, nil
	}
	switch t := spec.(type) {
	case string:
		return ParseHints(t, implicit...), nil
	case ReferenceHint:
		return ReferenceHints{t}, nil
	case []ReferenceHint:
		return t, nil
	case ReferenceHints:
		return t, nil
	case []DefaultReferenceHint:
		return sliceutils.Convert[ReferenceHint](t), nil
	case DefaultReferenceHints:
		return sliceutils.Convert[ReferenceHint](t), nil
	case map[string]string:
		return ReferenceHints{DefaultReferenceHint(t)}, nil
	case common.Properties:
		return ReferenceHints{DefaultReferenceHint(t)}, nil
	}
	return nil, fmt.Errorf("unknown hint specification type (%T)", spec)
}
