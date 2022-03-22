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
	"reflect"
)

type errRecursion struct {
	wrapped error
	kind    string
	elem    interface{}
	hist    []interface{}
}

// ErrRecusion describes a resursion errors caused by a dedicated element with an element history
func ErrRecusion(kind string, elem interface{}, hist interface{}) error {
	return &errRecursion{nil, kind, elem, ToInterfaceSlice(hist)}
}

func ErrRecusionWrap(err error, kind string, elem interface{}, hist interface{}) error {
	return &errRecursion{err, kind, elem, ToInterfaceSlice(hist)}
}

func (e *errRecursion) Error() string {
	msg := fmt.Sprintf("%s recursion: use of %v", e.kind, e.elem)
	if len(e.hist) > 0 {
		s := ""
		sep := ""
		for _, h := range e.hist {
			s = fmt.Sprintf("%s%s%v", s, sep, h)
			sep = "->"
		}
		msg = fmt.Sprintf("%s for %s", msg, s)
	}
	if e.wrapped != nil {
		return msg + ": " + e.wrapped.Error()
	}
	return msg
}

func (e *errRecursion) Unwrap() error {
	return e.wrapped
}

func (e *errRecursion) Elem() interface{} {
	return e.elem
}

func (e *errRecursion) Kind() string {
	return e.kind
}

func IsErrRecusion(err error) bool {
	return IsA(err, &errRecursion{})
}

func IsErrRecursionKind(err error, kind string) bool {
	var uerr *errRecursion
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}

func ToInterfaceSlice(list interface{}) []interface{} {
	if list == nil {
		return nil
	}
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		panic("no array or slice")
	}
	r := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		r[i] = v.Index(i).Interface()
	}
	return r
}
