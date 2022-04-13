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

package utils

import (
	"sort"
)

type StringSlice []string

func (l *StringSlice) Add(list ...string) {
	*l = append(*l, list...)
}

func (l *StringSlice) Delete(i int) {
	*l = append((*l)[:i], (*l)[i+1:]...)
}

func (l StringSlice) Contains(s string) bool {
	for _, e := range l {
		if e == s {
			return true
		}
	}
	return false
}
func (l StringSlice) Sort() {
	sort.Strings(l)
}
