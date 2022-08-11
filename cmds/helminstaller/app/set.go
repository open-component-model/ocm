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

package app

import (
	"fmt"
	"strings"
)

// TODO: support better path expressions

func Set(values map[string]interface{}, path string, value interface{}) error {
	fields := strings.Split(path, ".")
	i := 0
	for ; i < len(fields)-1; i++ {
		f := strings.TrimSpace(fields[i])
		v, ok := values[f]
		if !ok {
			v = map[string]interface{}{}
			values[f] = v
		} else {
			if _, ok := v.(map[string]interface{}); !ok {
				return fmt.Errorf("invalid field path %s", strings.Join(fields[:i+1], "."))
			}
		}
		values = v.(map[string]interface{})
	}
	values[fields[len(fields)-1]] = value
	return nil
}
