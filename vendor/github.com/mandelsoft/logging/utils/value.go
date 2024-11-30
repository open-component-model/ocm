/*
 * Copyright 2024 Mandelsoft. All rights reserved.
 *  This file is licensed under the Apache Software License, v. 2 except as noted
 *  otherwise in the LICENSE file
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// use package for package doc
var (
	_ logr.Logger
	_ logrus.Logger
)

type ignoreKeyPair struct{}

func (n ignoreKeyPair) String() string {
	return "<unset>"
}

// Ignore can be used as value for a message field, which should not be present,
// but is still given in the values list. This can be used to avoid complex value
// list generation for optional fields of a logging call.
//
// This value is only considered as special value for the logging functions
// of this module, or if a message formatter of this module is used.
// For plain [logrus] formatter, or [logr] logger it is just a regular value.
// Its string representation is "<unset>"-
var Ignore = &ignoreKeyPair{}

func FieldValue(formatter func(interface{}) string, v interface{}) interface{} {
	if v == Ignore {
		return v
	}
	// Try to avoid marshaling known types.
	switch vVal := v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, complex64, complex128,
		string, bool:
		return vVal

	case []byte:
		return string(vVal)
	case fmt.Stringer:
		return vVal.String()
	case error:
		return vVal.Error()

	default:
		vv := reflect.ValueOf(v)
		switch vv.Kind() {
		case reflect.Struct, reflect.Map:
			if formatter == nil {
				j, _ := json.Marshal(vVal)
				return string(j)
			}
			fallthrough
		default:
			if formatter != nil {
				return formatter(v)
			}
			return v
		}
	}
}
