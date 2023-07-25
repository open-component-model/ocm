// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package tree

import (
	"github.com/open-component-model/ocm/v2/cmds/ocm/pkg/data"
)

type Objects []Object

func ObjectSlice(s data.Iterable) Objects {
	var a Objects
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().(Object))
	}
	return a
}

var (
	_ data.IndexedAccess = Objects{}
	_ data.Iterable      = Objects{}
)

func (o Objects) Len() int {
	return len(o)
}

func (o Objects) Get(i int) interface{} {
	return o[i]
}

func (o Objects) Iterator() data.Iterator {
	return data.NewIndexedIterator(o)
}

////////////////////////////////////////////////////////////////////////////////

type TreeObjects []*TreeObject

var (
	_ data.IndexedAccess = TreeObjects{}
	_ data.Iterable      = TreeObjects{}
)

func (o TreeObjects) Len() int {
	return len(o)
}

func (o TreeObjects) Get(i int) interface{} {
	return o[i]
}

func (o TreeObjects) Iterator() data.Iterator {
	return data.NewIndexedIterator(o)
}
