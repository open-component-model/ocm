// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type NoMatchError struct {
	errinfo
}

var formatNoMatch = NewDefaultFormatter("has", "no match", "in")

func ErrNoMatch(spec ...string) error {
	return &InvalidError{newErrInfo(formatNoMatch, spec...)}
}

func ErrNoMatchWrap(err error, spec ...string) error {
	return &InvalidError{wrapErrInfo(err, formatNoMatch, spec...)}
}

func IsErrNoMatch(err error) bool {
	return IsA(err, &InvalidError{})
}

func IsErrNoMatchKind(err error, kind string) bool {
	var uerr *InvalidError
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
