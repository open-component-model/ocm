// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"sort"
)

type CompareFunction func(interface{}, interface{}) int

type elements struct {
	data    []interface{}
	compare CompareFunction
}

func (a *elements) Len() int           { return len(a.data) }
func (a *elements) Swap(i, j int)      { a.data[i], a.data[j] = a.data[j], a.data[i] }
func (a *elements) Less(i, j int) bool { return a.compare(a.data[i], a.data[j]) < 0 }

func Sort(data []interface{}, cmp CompareFunction) {
	sort.Sort(&elements{data, cmp})
}

type CompareIndexedFunction func(int, interface{}, int, interface{}) int

type indexed struct {
	data    []interface{}
	compare CompareIndexedFunction
}

func (a *indexed) Len() int           { return len(a.data) }
func (a *indexed) Swap(i, j int)      { a.data[i], a.data[j] = a.data[j], a.data[i] }
func (a *indexed) Less(i, j int) bool { return a.compare(i, a.data[i], j, a.data[j]) < 0 }

func SortIndexed(data []interface{}, cmp CompareIndexedFunction) {
	sort.Sort(&indexed{data, cmp})
}

type view struct {
	data    []interface{}
	view    []int
	compare CompareFunction
}

func (a *view) Len() int { return len(a.view) }
func (a *view) Swap(i, j int) {
	a.data[a.view[i]], a.data[a.view[j]] = a.data[a.view[j]], a.data[a.view[i]]
}
func (a *view) Less(i, j int) bool { return a.compare(a.data[a.view[i]], a.data[a.view[j]]) < 0 }

func SortView(data []interface{}, mapping []int, cmp CompareFunction) {
	if len(mapping) > 1 {
		sort.Sort(&view{data, mapping, cmp})
	}
}
