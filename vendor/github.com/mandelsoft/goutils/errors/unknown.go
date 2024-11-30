package errors

type UnknownError struct {
	errinfo
}

var formatUnknown = NewDefaultFormatter("is", "unknown", "for")

func ErrUnknown(spec ...string) error {
	return &UnknownError{newErrInfo(formatUnknown, spec...)}
}

func IsErrUnknown(err error) bool {
	return IsA(err, &UnknownError{})
}

func IsErrUnknownKind(err error, kind string) bool {
	var uerr *UnknownError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
