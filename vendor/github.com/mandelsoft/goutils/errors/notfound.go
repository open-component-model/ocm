package errors

type NotFoundError struct {
	errinfo
}

var formatNotFound = NewDefaultFormatter("", "not found", "in")

func ErrNotFound(spec ...string) error {
	return &NotFoundError{newErrInfo(formatNotFound, spec...)}
}

func ErrNotFoundWrap(err error, spec ...string) error {
	return &NotFoundError{wrapErrInfo(err, formatNotFound, spec...)}
}

func IsErrNotFound(err error) bool {
	return IsA(err, &NotFoundError{})
}

func IsErrNotFoundKind(err error, kind string) bool {
	var uerr *NotFoundError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}

func IsErrNotFoundElem(err error, kind, elem string) bool {
	var uerr *NotFoundError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind && uerr.elem != nil && *uerr.elem == elem
}
