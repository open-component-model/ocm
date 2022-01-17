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
	"errors"
	"fmt"
	"reflect"
)

var New = errors.New
var Unwrap = errors.Unwrap
var Is = errors.Is
var As = errors.As

func Newf(msg string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(msg, args...))
}

func IsA(err error, target error) bool {
	if err == nil {
		return false
	}
	typ := reflect.TypeOf(target)

	for err != nil {
		if reflect.TypeOf(err).AssignableTo(typ) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////

type wrappedError struct {
	parent error
	msg    string
}

func Wrapf(err error, msg string, args ...interface{}) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &wrappedError{
		parent: err,
		msg:    msg,
	}
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %s", e.msg, e.parent)
}

func (e *wrappedError) Unwrap() error {
	return e.parent
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

////////////////////////////////////////////////////////////////////////////////

type errinfo struct {
	msg         string
	preposition string
	kind        string
	elem        string
	ctx         string
}

func newErrInfo(msg, preposition string, spec ...string) errinfo {
	e := errinfo{
		msg:         msg,
		preposition: preposition,
	}

	if len(spec) > 2 {
		e.kind = spec[0]
		e.elem = spec[1]
		e.ctx = spec[2]
		return e
	}
	if len(spec) > 1 {
		e.kind = spec[0]
		e.elem = spec[1]
		return e
	}
	if len(spec) > 0 {
		e.elem = spec[0]
	}
	return e
}

func (e *errinfo) Error() string {
	ctx := ""
	if e.ctx != "" {
		ctx = " " + e.preposition + " " + e.ctx
	}
	elem := ""
	if e.elem != "" {
		elem = e.elem + " "
	}
	kind := ""
	if e.kind != "" {
		kind = e.kind + " "
	}
	return kind + elem + e.msg + ctx
}

func (e *errinfo) Elem() string {
	return e.elem
}

func (e *errinfo) Kind() string {
	return e.kind
}

func (e *errinfo) Ctx() string {
	return e.ctx
}
