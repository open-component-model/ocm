package errors

import (
	"fmt"
)

type InvalidError struct {
	errinfo
}

var formatInvalid = NewDefaultFormatter("is", "invalid", "for")

func ErrInvalid(spec ...string) error {
	return &InvalidError{newErrInfo(formatInvalid, spec...)}
}

// ErrInvalidType reports an invalid or unexpected Go type for a dedicated purpose.
func ErrInvalidType(kind string, v interface{}) error {
	return &InvalidError{newErrInfo(formatUnknown, kind, fmt.Sprintf("%T", v))}
}

func ErrInvalidWrap(err error, spec ...string) error {
	return &InvalidError{wrapErrInfo(err, formatInvalid, spec...)}
}

func IsErrInvalid(err error) bool {
	return IsA(err, &InvalidError{})
}

func IsErrInvalidKind(err error, kind string) bool {
	var uerr *InvalidError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
