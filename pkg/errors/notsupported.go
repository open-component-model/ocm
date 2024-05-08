package errors

type NotSupportedError struct {
	errinfo
}

var formatNotSupported = NewDefaultFormatter("", "not supported", "by")

func ErrNotSupported(spec ...string) error {
	return &NotSupportedError{newErrInfo(formatNotSupported, spec...)}
}

func IsErrNotSupported(err error) bool {
	return IsA(err, &NotSupportedError{})
}

func IsErrNotSupportedKind(err error, kind string) bool {
	var uerr *NotSupportedError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
