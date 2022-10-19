// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type errUnknown struct {
	errinfo
}

var formatUnknown = NewDefaultFormatter("is", "unknown", "for")

func ErrUnknown(spec ...string) error {
	return &errUnknown{newErrInfo(formatUnknown, spec...)}
}

func IsErrUnknown(err error) bool {
	return IsA(err, &errUnknown{})
}

func IsErrUnknownKind(err error, kind string) bool {
	var uerr *errUnknown
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
