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

package template

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/drone/envsubst"
)

type Subst struct{}

var _ Templater = (*Subst)(nil)

func NewSubst() Templater {
	return &Subst{}
}

// Template templates a string with the parsed vars.
func (s *Subst) Process(data string, values Values) (string, error) {
	return envsubst.Eval(data, stringmapping(values))
}

// mapping is a helper function for the envsubst to provide the value for a variable name.
// It returns an empty string if the variable is not defined.
func stringmapping(values Values) func(variable string) string {
	return func(variable string) string {
		if values == nil {
			return ""
		}
		v := values[variable]
		if v == nil {
			return ""
		}
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Map || t.Kind() == reflect.Array {
			data, err := json.Marshal(v)
			if err != nil {
				return ""
			}
			return string(data)
		}
		return fmt.Sprintf("%v", v)
	}
}
