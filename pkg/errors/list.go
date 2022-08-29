// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package errors

import (
	"fmt"

	"github.com/open-component-model/ocm/pkg/common/printer"
)

type ErrorList struct {
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

func (l *ErrorList) Addf(pr printer.Printer, err error, msg string, args ...interface{}) error {
	if err != nil {
		if msg != "" {
			err = Wrapf(err, msg, args...)
		}
		l.errors = append(l.errors, err)
		if pr != nil {
			pr.Errorf("%s\n", err)
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
