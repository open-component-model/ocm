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

package main

import (
	"fmt"

	"github.com/gardener/ocm/cmds/ocm/pkg/tree"
	"github.com/gardener/ocm/pkg/common"
)

type Elem struct {
	History common.History
	Node    bool
	Data    string
}

var _ tree.Object = (*Elem)(nil)

func (e *Elem) GetHierarchy() common.History {
	return e.History
}

func (e *Elem) IsNode() *common.NameVersion {
	if e.Node {
		n := common.NewNameVersion(e.Data, "")
		return &n
	}
	return nil
}

func (e *Elem) String() string {
	return e.Data
}

func E(d string, hist ...string) *Elem {
	h := common.History{}
	for _, v := range hist {
		h = append(h, common.NewNameVersion(v, ""))
	}
	return &Elem{h, false, d}
}

func N(d string, hist ...string) *Elem {
	h := common.History{}
	for _, v := range hist {
		h = append(h, common.NewNameVersion(v, ""))
	}
	return &Elem{h, true, d}
}

func Create(h common.History, n common.NameVersion) tree.Object {
	return &Elem{h, true, n.GetName()}
}

func main() {
	data := []tree.Object{
		E("a"),
		N("b"),
		E("a", "b"),
		E("a", "b", "c"),
		E("a", "e", "f"),
		E("c"),
		E("d"),
	}

	t := tree.MapToTree(data, Create)
	for _, l := range t {
		fmt.Printf("%s %s\n", l.Graph, l.Object)
	}
}
