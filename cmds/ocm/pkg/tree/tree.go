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

package tree

import (
	"github.com/gardener/ocm/pkg/common"
)

type Element interface {
	GetHierarchy() common.History
	IsNode() *common.NameVersion
}

type NodeCreator func(common.History, common.NameVersion) Element

type Elements []Element

type TreeElement struct {
	Header  string
	Element Element
}

var vertical = "│" + space[1:]
var horizontal = "─"
var corner = "└" + horizontal
var fork = "├" + horizontal
var space = "   "

func BuildTree(elems Elements, creator NodeCreator) []TreeElement {
	result := []TreeElement{}
	handleLevel(elems, "", nil, 0, creator, &result)
	return result
}

func handleLevel(elems Elements, header string, prefix common.History, start int, creator NodeCreator, result *[]TreeElement) {
	var node *common.NameVersion
	lvl := len(prefix)
	for i := start; i < len(elems); {
		var next int
		h := elems[i].GetHierarchy()
		if !h.HasPrefix(prefix) {
			return
		}
		ftag := corner
		stag := space
		for next = i + 1; next < len(elems); next++ {
			if s := elems[next].GetHierarchy(); s.HasPrefix(prefix) {
				if len(s) > lvl && len(h) > lvl && h[lvl] == s[lvl] { // skip same sub level
					continue
				}
				ftag = fork
				stag = vertical
			}
			break
		}
		if len(h) == lvl {
			node = elems[i].IsNode() // Element acts as dedicate node
			*result = append(*result, TreeElement{
				Header:  header + ftag,
				Element: elems[i],
			})
			i++
		} else {
			if node == nil || *node != h[lvl] {
				// synthesize node if only leafs or non-matching node has been issued before
				*result = append(*result, TreeElement{
					Header:  header + ftag, // + " " + h[len(prefix)].String(),
					Element: creator(prefix, h[len(prefix)]),
				})
			}
			handleLevel(elems, header+stag, h[:len(prefix)+1], i, creator, result)
			i = next
			node = nil
		}
	}
}
