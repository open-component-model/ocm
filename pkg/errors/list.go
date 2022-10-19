// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"io"
)

// ErrorList is an error type with erros in it.
type ErrorList struct { //nolint: errname // Intentional naming.
	msg    string
	errors []error
}

func (l *ErrorList) Error() string {
	msg := "{" + l.msg
	sep := ""
	if msg != "" {
		sep = ": "
	}
	for _, e := range l.errors {
		if e != nil {
			msg = fmt.Sprintf("%s%s%s", msg, sep, e)
			sep = ", "
		}
	}
	return msg + "}"
}

func (l *ErrorList) Add(errs ...error) *ErrorList {
	for _, e := range errs {
		if e != nil {
			l.errors = append(l.errors, e)
		}
	}
	return l
}

func (l *ErrorList) Addf(writer io.Writer, err error, msg string, args ...interface{}) error {
	if err != nil {
		if msg != "" {
			err = Wrapf(err, msg, args...)
		}
		l.errors = append(l.errors, err)
		if writer != nil {
			fmt.Fprintf(writer, "Error: %s\n", err)
		}
	}
	return err
}

func (l *ErrorList) Len() int {
	return len(l.errors)
}

func (l *ErrorList) Result() error {
	if l == nil || len(l.errors) == 0 {
		return nil
	}
	return l
}

func (l *ErrorList) Clear() {
	l.errors = nil
}

func ErrListf(msg string, args ...interface{}) *ErrorList {
	return &ErrorList{
		msg: fmt.Sprintf(msg, args...),
	}
}
