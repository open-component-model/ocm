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

type errNotFound struct {
	errinfo
}

var formatNotFound = NewDefaultFormatter("", "not found", "in")

func ErrNotFound(spec ...string) error {
	return &errNotFound{newErrInfo(formatNotFound, spec...)}
}

func ErrNotFoundWrap(err error, spec ...string) error {
	return &errNotFound{wrapErrInfo(err, formatNotFound, spec...)}
}

func IsErrNotFound(err error) bool {
	return IsA(err, &errNotFound{})
}

func IsErrNotFoundKind(err error, kind string) bool {
	var uerr *errNotFound
	if err == nil || !As(err, &uerr) {
		return false
	}
	return uerr.kind == kind
}
