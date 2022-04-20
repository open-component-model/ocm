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
	"fmt"

	"github.com/open-component-model/ocm/pkg/common"
)

type Object interface {
	common.HistorySource
	IsNode() *common.NameVersion
}

type Typed interface {
	GetKind() string
}

type NodeCreator func(common.History, common.NameVersion) Object

// TreeObject is an element enriched by a textual
// tree graph prefix line
type TreeObject struct {
	Graph  string
	Object Object
	Node   *TreeNode // for syntesized noded this entry is used if no object can be synthesized
}

func (t *TreeObject) String() string {
	if t.Object != nil {
		return fmt.Sprintf("%s %s", t.Graph, t.Object)
	}
	return fmt.Sprintf("%s %s", t.Graph, t.Node.String())
}

type TreeNode struct {
	common.NameVersion
	History common.History
}

var vertical = "│" + space[1:]
var horizontal = "─"
var corner = "└" + horizontal
var fork = "├" + horizontal
var space = "   "

// MapToTree maps a list of elements featuring a resulution history
// into a list of elements providing an ascii tree graph field
func MapToTree(objs Objects, creator NodeCreator) TreeObjects {
	result := TreeObjects{}
	handleLevel(objs, "", nil, 0, creator, &result)
	return result
}

func handleLevel(objs Objects, header string, prefix common.History, start int, creator NodeCreator, result *TreeObjects) {
	var node *common.NameVersion
	lvl := len(prefix)
	for i := start; i < len(objs); {
		var next int
		h := objs[i].GetHistory()
		if !h.HasPrefix(prefix) {
			return
		}
		ftag := corner
		stag := space
		key := objs[i].IsNode()
		for next = i + 1; next < len(objs); next++ {
			if s := objs[next].GetHistory(); s.HasPrefix(prefix) {
				if len(s) > lvl && len(h) > lvl && h[lvl] == s[lvl] { // skip same sub level
					continue
				}
				if key != nil {
					if len(s) > lvl && *key == s[lvl] { // skip same sub level
						continue
					}
				}
				ftag = fork
				stag = vertical
			}
			break
		}
		if len(h) == lvl {
			node = objs[i].IsNode() // Element acts as dedicate node
			sym := ""
			if node != nil {
				if i < len(objs)-1 {
					sub := objs[i+1].GetHistory()
					if len(sub) > len(h) && sub.HasPrefix(append(h, *node)) {
						sym = " \u2297"
					}
				}
			}
			if t, ok := objs[i].(Typed); ok {
				k := t.GetKind()
				if k != "" {
					sym += " " + k
				}
			}
			*result = append(*result, &TreeObject{
				Graph:  header + ftag + sym,
				Object: objs[i],
			})
			i++
		} else {
			if node == nil || *node != h[lvl] {
				// synthesize node if only leafs or non-matching node has been issued before
				var o Object
				var n *TreeNode
				if creator != nil {
					o = creator(prefix, h[len(prefix)])
				}
				if o == nil {
					n = &TreeNode{h[len(prefix)], prefix}
				}
				*result = append(*result, &TreeObject{
					Graph:  header + ftag, // + " " + h[len(prefix)].String(),
					Object: o,
					Node:   n,
				})
			}
			handleLevel(objs, header+stag, h[:len(prefix)+1], i, creator, result)
			i = next
			node = nil
		}
	}
}
