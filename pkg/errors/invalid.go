// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type errInvalid struct {
	errinfo
}

var formatInvalid = NewDefaultFormatter("is", "invalid", "for")

func ErrInvalid(spec ...string) error {
	return &errInvalid{newErrInfo(formatInvalid, spec...)}
}

func ErrInvalidWrap(err error, spec ...string) error {
	return &errInvalid{wrapErrInfo(err, formatInvalid, spec...)}
}

func IsErrInvalid(err error) bool {
	return IsA(err, &errInvalid{})
}

func IsErrInvalidKind(err error, kind string) bool {
	var uerr *errInvalid
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
