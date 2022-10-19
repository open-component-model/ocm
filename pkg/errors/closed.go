// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type errClosed struct {
	errinfo
}

var formatClosed = NewDefaultFormatter("is", "closed", "for")

func ErrClosed(spec ...string) error {
	return &errClosed{newErrInfo(formatClosed, spec...)}
}

func IsErrClosed(err error) bool {
	return IsA(err, &errClosed{})
}

func IsErrClosedKind(err error, kind string) bool {
	var uerr *errClosed
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
