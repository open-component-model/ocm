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

package data

import (
	"reflect"

	"github.com/modern-go/reflect2"
)

func IsNil(i interface{}) bool {
	return reflect2.IsNil(i)
}

func IsEmpty(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.ValueOf(i).Kind() {
	case reflect.Map | reflect.Array | reflect.Slice | reflect.String:
		return reflect.ValueOf(i).Len() == 0
	case reflect.Ptr | reflect.Interface:
		if reflect.ValueOf(i).IsNil() {
			return true
		}
		return IsEmpty(reflect.ValueOf(i).Elem().Interface())
	}
	return false
}
