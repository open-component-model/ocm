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
	"fmt"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

type StringElementList []string

func (l StringElementList) Size() int                { return len(l) }
func (l StringElementList) Key(i int) string         { return l[i] }
func (l StringElementList) Description(i int) string { return "" }

func FormatList(def string, elems ...string) string {
	return FormatListElements(def, StringElementList(elems))
}

type ListElements interface {
	Size() int
	Key(i int) string
	Description(i int) string
}

func FormatListElements(def string, elems ListElements) string {
	names := ""
	size := elems.Size()

	for i := 0; i < size; i++ {
		key := elems.Key(i)
		names = fmt.Sprintf("%s\n  - <code>%s</code>", names, key)
		if key == def {
			names += " (default)"
		}
		desc := elems.Description(i)
		if len(desc) > 0 {
			names += ": " + desc
		}
	}
	return names + "\n"
}

type IdentityMatcherList []credentials.IdentityMatcherInfo

func (l IdentityMatcherList) Size() int                { return len(l) }
func (l IdentityMatcherList) Key(i int) string         { return l[i].Type }
func (l IdentityMatcherList) Description(i int) string { return l[i].Description }
