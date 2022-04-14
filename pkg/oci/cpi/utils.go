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

package cpi

import (
	"strings"

	"github.com/open-component-model/ocm/pkg/oci/grammar"
)

type StringList []string

func (s *StringList) Add(n string) {
	for _, e := range *s {
		if n == e {
			return
		}
	}
	*s = append(*s, n)
}

func FilterByNamespacePrefix(prefix string, list []string) []string {
	result := []string{}
	sub := prefix
	if prefix != "" && !strings.HasSuffix(prefix, grammar.RepositorySeparator) {
		sub = prefix + grammar.RepositorySeparator
	}
	for _, k := range list {
		if k == prefix || strings.HasPrefix(k, sub) {
			result = append(result, k)
		}
	}
	return result
}

func FilterChildren(closure bool, list []string) []string {
	if closure {
		return list
	}
	set := map[string]bool{}
	for _, n := range list {
		i := strings.Index(n, grammar.RepositorySeparator)
		if i < 0 {
			set[n] = true
		} else {
			set[n[:i]] = true
		}
	}
	result := make([]string, 0, len(set))
	for n := range set {
		result = append(result, n)
	}
	return result
}
