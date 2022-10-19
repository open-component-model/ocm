// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package data

func Slice(s Iterable) []interface{} {
	var a []interface{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next())
	}
	return a
}

func StringArraySlice(s Iterable) [][]string {
	a := [][]string{}
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().([]string))
	}
	return a
}
