// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

type errReadOnly struct {
	errinfo
}

var formatReadOnly = NewDefaultFormatter("is", "readonly", "in")

func ErrReadOnly(spec ...string) error {
	return &errReadOnly{newErrInfo(formatReadOnly, spec...)}
}

func IsErrReadOnly(err error) bool {
	return IsA(err, &errReadOnly{})
}

func IsErrReadOnlyKind(err error, kind string) bool {
	var uerr *errReadOnly
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
