// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type errNotImplemented struct {
	errinfo
}

var formatNotImplemented = NewDefaultFormatter("", "not implemented", "by")

func ErrNotImplemented(spec ...string) error {
	return &errNotImplemented{newErrInfo(formatNotImplemented, spec...)}
}

func IsErrNotImplemented(err error) bool {
	return IsA(err, &errNotImplemented{})
}

func IsErrNotImplementedKind(err error, kind string) bool {
	var uerr *errNotImplemented
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
