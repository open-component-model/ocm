package refhints

import (
	"github.com/mandelsoft/goutils/general"
)

type state int

const (
	sStart state = iota
	sTypeOrValue
	sColonInType
	sPropStart
	sProp
	sValueStart
	sPlainValue
	sEscapedValue
	sEscapedEnd
)

// ParseHints parses a string containing servialized reference hints,
// If implicit is set to true, the implicit attribute is set.
//
// In general a hint is serialized to the following string:
//
//	[<*type*>`::]`<*attribute*>`=`<*value*>{`,`<*attribute*>`=`<*value*>}
//
// The type is not serialized as attribute, but as prefix separated by a ::.
// The implicit attribute is never serialized if the string is stored in an
// access specification.
// If no type is known the type part is omitted.
//
// A list of hints is serialized to
//
//	<*hint*>{`;`<*hint*>}
//
// Attributes names consist of alphanumeric characters, only.
// A value (or type) may not contain a ::. If it contains a ;, , or "
// character it must be given in double quotes.
// In the double-quoted form any " or \ character has to be escaped by
// a preceding \ character.
//
// To be as compatible as possible, a single attribute hint with the attribute
// reference is serialized as naked value (as before) if there are no special
// characters enforcing a quoted form.
//
// see DefaultReferenceHint.Serialize.
func ParseHints(v string, implicit ...bool) ReferenceHints {
	var hints ReferenceHints

	var prop string
	var val string

	var hint DefaultReferenceHint
	state := sStart
	start := 0
	mask := false
	impl := general.Optional(implicit...)
	for i, c := range v {
		switch state {
		case sStart:
			if c == '"' {
				hint = newHint(impl)
				prop = HINT_REFERENCE
				start = i + 1
				state = sEscapedValue
			} else {
				state = sTypeOrValue
			}
			fallthrough
		case sTypeOrValue: // type or plain value
			if c == ':' {
				state = sColonInType
			}
			if c == '=' {
				hint = DefaultReferenceHint{}
				prop = v[start:i]
				start = i + 1
				state = sValueStart
			}
			if c == ',' || c == ';' {
				hint = DefaultReferenceHint{}
				hint[HINT_REFERENCE] = v[start:i]
				start = i + 1
				if c == ',' {
					state = sPropStart
				}
				if c == ';' {
					hints = append(hints, hint)
					hint = nil
					state = sStart
				}
			}
		case sColonInType: // colon
			if c == ':' {
				hint = newHint(impl).SetProperty(HINT_TYPE, v[start:i-1])
				start = i + 1
				state = sPropStart
			} else {
				state = sTypeOrValue
			}
		case sPropStart: // prop start
			if c == '"' {
				val = ""
				prop = HINT_REFERENCE
				state = sEscapedValue
				start = i + 1
				continue
			}
			state = sProp
			fallthrough
		case sProp: // prop
			switch c {
			case '=':
				prop = v[start:i]
				start = i + 1
				state = sValueStart
			case ';':
				hint[HINT_REFERENCE] = v[start:i]
				hints = append(hints, hint)
				hint = nil
				state = sStart
				start = i + 1
			}
		case sValueStart: // value start
			if c == '"' {
				val = ""
				state = sEscapedValue
				start = i + 1
			} else {
				state = sPlainValue
				start = i
			}
		case sPlainValue: // plain value
			if c == ',' || c == ';' {
				hint[prop] = v[start:i]
				start = i + 1
				if c == ';' {
					hints = append(hints, hint)
					hint = nil
					state = sStart
				} else {
					state = sPropStart
				}
			}
		case sEscapedValue: // escaped value
			if mask {
				mask = false
			} else {
				if c == '\\' {
					mask = true
					continue
				}
				if c == '"' {
					hint[prop] = val
					state = sEscapedEnd
				}
			}
			val += string(c)
		case sEscapedEnd: // end escaped
			if c == ',' {
				start = i + 1
				state = sProp
			}
			if c == ';' {
				hints = append(hints, hint)
				hint = nil
				start = i + 1
				state = sStart
			}
		}
	}

	switch state {
	case sTypeOrValue, sColonInType:
		hint = newHint(impl).SetProperty(HINT_REFERENCE, v[start:])
	case sProp:
		hint[HINT_REFERENCE] = v[start:]
	case sValueStart:
		hint[prop] = ""
	case sPlainValue:
		hint[prop] = v[start:]
	case sEscapedValue:
		hint[prop] = v[start:]
	case sPropStart:
	case sEscapedEnd:
	}
	hints = append(hints, hint)
	return hints
}
