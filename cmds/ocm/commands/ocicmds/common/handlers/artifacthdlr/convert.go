package artifacthdlr

import (
	"ocm.software/ocm/cmds/ocm/common/data"
)

type Objects []*Object

func ObjectSlice(s data.Iterable) Objects {
	var a Objects
	i := s.Iterator()
	for i.HasNext() {
		a = append(a, i.Next().(*Object))
	}
	return a
}

var (
	_ data.IndexedAccess = Objects{}
	_ data.Iterable      = Objects{}
)

func (this Objects) Len() int {
	return len(this)
}

func (this Objects) Get(i int) interface{} {
	return this[i]
}

func (this Objects) Iterator() data.Iterator {
	return data.NewIndexedIterator(this)
}
