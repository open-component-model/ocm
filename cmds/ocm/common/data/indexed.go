package data

type IndexedAccess interface {
	Len() int
	Get(int) interface{}
}

type IndexedIterator struct {
	access  IndexedAccess
	current int
}

var _ ResettableIterator = &IndexedIterator{}

func NewIndexedIterator(a IndexedAccess) *IndexedIterator {
	return (&IndexedIterator{}).new(a)
}

func (this *IndexedIterator) new(a IndexedAccess) *IndexedIterator {
	this.access = a
	this.current = -1
	return this
}

func (this *IndexedIterator) HasNext() bool {
	return this.access.Len() > this.current+1
}

func (this *IndexedIterator) Next() interface{} {
	if this.HasNext() {
		this.current++
		return this.access.Get(this.current)
	}
	return nil
}

func (this *IndexedIterator) Reset() {
	this.current = -1
}

////////////////////////////////////////////////////////////////////////////

type IndexedSliceAccess []interface{}

var (
	_ IndexedAccess = IndexedSliceAccess{}
	_ Iterable      = IndexedSliceAccess{}
)

func (this *IndexedSliceAccess) Add(elems ...interface{}) *IndexedSliceAccess {
	*this = append(*this, elems...)
	return this
}

func (this *IndexedSliceAccess) Remove(i int) *IndexedSliceAccess {
	*this = append((*this)[:i], (*this)[i+1:]...)
	return this
}

// Move [i:j] to [k:].
func (this *IndexedSliceAccess) Move(i, j, k int) *IndexedSliceAccess {
	if k < i || k > j {
		l := j - i

		n := make(IndexedSliceAccess, len(*this))
		if i > k {
			copy(n[k:], (*this)[i:j])
			copy(n[k+l:], (*this)[k:i])
			copy((*this)[k:], n[k:j])
		} else {
			copy(n[i:], (*this)[j:k])
			copy(n[i+k-j:], (*this)[i:j])
			copy((*this)[i:], n[i:k])
		}
	}
	return this
}

func (this IndexedSliceAccess) Len() int {
	return len(this)
}

func (this IndexedSliceAccess) Get(i int) interface{} {
	return this[i]
}

func (this IndexedSliceAccess) Iterator() Iterator {
	return NewIndexedIterator(this)
}

func (this IndexedSliceAccess) Sort(cmp CompareFunction) IndexedSliceAccess {
	Sort(this, cmp)
	return this
}

func (this IndexedSliceAccess) SortIndexed(cmp CompareIndexedFunction) IndexedSliceAccess {
	SortIndexed(this, cmp)
	return this
}

func (this IndexedSliceAccess) Copy() IndexedSliceAccess {
	n := make(IndexedSliceAccess, len(this))
	copy(n, this)
	return n
}

func NewSliceIterator(slice []interface{}) *IndexedIterator {
	return NewIndexedIterator(IndexedSliceAccess(slice))
}
