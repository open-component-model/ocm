package tree

import (
	"ocm.software/ocm/cmds/ocm/common/data"
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
