package data

type Iterable interface {
	Iterator() Iterator
}

type Iterator interface {
	HasNext() bool
	Next() interface{}
}

type ResettableIterator interface {
	HasNext() bool
	Next() interface{}
	Reset()
}

type MappingFunction func(interface{}) interface{}

type MappedIterator struct {
	Iterator
	mapping MappingFunction
}

func NewMappedIterator(iter Iterator, mapping MappingFunction) Iterator {
	return &MappedIterator{iter, mapping}
}

func (mi *MappedIterator) Next() interface{} {
	if mi.HasNext() {
		return mi.mapping(mi.Iterator.Next())
	}
	return nil
}
