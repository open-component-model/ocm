package processing

import (
	"sync"

	"github.com/mandelsoft/logging"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocmlog "ocm.software/ocm/api/utils/logging"
	"ocm.software/ocm/cmds/ocm/common/data"
)

type Result struct {
	result []interface{}
	wg     sync.WaitGroup
}

func Gather(i data.Iterable) *Result {
	r := &Result{}
	r.wg.Add(1)
	go func() {
		// defer GinkgoRecover()
		r.result = data.Slice(i)
		r.wg.Done()
	}()
	return r
}

func (r *Result) Get() []interface{} {
	r.wg.Wait()
	return r.result
}

func ExpectNext(it data.Iterator, v int, next bool) {
	if !it.(CheckNext).CheckNext() {
		Fail("next element expected but not indicated", 1)
	}
	ExpectWithOffset(1, it.Next()).To(Equal(v))
	ExpectWithOffset(1, it.(CheckNext).CheckNext()).To(Equal(next))
}

var _ = Describe("processing buffer", func() {
	var log logging.Context

	BeforeEach(func() {
		log, _ = ocmlog.NewBufferedContext()
	})

	Context("index array", func() {
		It("after empty", func() {
			i := IndexArray{}
			Expect(i.After(IndexArray{1, 2, 3})).To(BeFalse())
		})
		It("after same level", func() {
			i := IndexArray{1, 2, 3}
			Expect(i.After(IndexArray{1, 2, 3})).To(BeFalse())
			Expect(i.After(IndexArray{1, 2, 4})).To(BeFalse())
			Expect(i.After(IndexArray{1, 3, 3})).To(BeFalse())
			Expect(i.After(IndexArray{2, 2, 3})).To(BeFalse())

			Expect(i.After(IndexArray{1, 2, 2})).To(BeTrue())
			Expect(i.After(IndexArray{1, 1, 3})).To(BeTrue())
			Expect(i.After(IndexArray{0, 2, 3})).To(BeTrue())
		})
		It("after deeper level", func() {
			i := IndexArray{1, 2, 3}
			Expect(i.After(IndexArray{1, 2, 3, 1})).To(BeFalse())
			Expect(i.After(IndexArray{1, 2, 2, 1})).To(BeTrue())
		})
		It("after shallower level", func() {
			i := IndexArray{1, 2, 3}
			Expect(i.After(IndexArray{1, 2})).To(BeTrue())
			Expect(i.After(IndexArray{1, 3})).To(BeFalse())
			Expect(i.After(IndexArray{1})).To(BeTrue())
			Expect(i.After(IndexArray{2, 2})).To(BeFalse())
			Expect(i.After(IndexArray{1, 1})).To(BeTrue())
		})
	})
	Context("index array next", func() {
		It("empty", func() {
			i := IndexArray{}
			Expect(i.Next(-1, 0)).To(Equal(IndexArray{0}))
		})
		It("down", func() {
			i := IndexArray{1}
			Expect(i.Next(-1, 0)).To(Equal(IndexArray{2}))
			Expect(i.Next(-1, 2)).To(Equal(IndexArray{1, 0}))
			Expect(i.Next(3, 0)).To(Equal(IndexArray{2}))
			Expect(i.Next(3, 2)).To(Equal(IndexArray{1, 0}))
		})
		It("up", func() {
			i := IndexArray{1, 2}
			Expect(i.Next(3, 0)).To(Equal(IndexArray{2}))
			Expect(i.Next(3, 2)).To(Equal(IndexArray{1, 2, 0}))
		})
	})

	Context("simple", func() {
		It("add", func() {
			buf := NewSimpleBuffer(log)

			promise := Gather(buf)

			buf.Add(NewEntry(Top(0), 0))
			buf.Add(NewEntry(Top(1), 1))
			buf.Add(NewEntry(Top(2), 2))
			buf.Add(NewEntry(Top(3), 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 1, 2, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{0, 1, 2, 3}))
		})
		It("add filtered", func() {
			buf := NewSimpleBuffer(log)

			promise := Gather(buf)

			buf.Add(NewEntry(Top(0), 0))
			buf.Add(NewEntry(Top(1), 1))
			buf.Add(NewEntry(Top(2), 2, false))
			buf.Add(NewEntry(Top(3), 3))
			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 1, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{0, 1, 2, 3}))
		})
	})

	Context("add ordered", func() {
		It("add in order", func() {
			buf := NewOrderedBuffer(log)

			promise := Gather(buf)
			it := buf.Iterator()
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(0), 0))
			ExpectNext(it, 0, false)

			buf.Add(NewEntry(Top(1), 1))
			ExpectNext(it, 1, false)

			buf.Add(NewEntry(Top(2), 2))
			ExpectNext(it, 2, false)

			buf.Add(NewEntry(Top(3), 3))
			ExpectNext(it, 3, false)

			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 1, 2, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{0, 1, 2, 3}))
		})

		It("add filtered", func() {
			buf := NewOrderedBuffer(log)

			promise := Gather(buf)
			it := buf.Iterator()
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(0), 0))
			ExpectNext(it, 0, false)

			buf.Add(NewEntry(Top(1), 1))
			ExpectNext(it, 1, false)

			buf.Add(NewEntry(Top(2), 2, false))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(3), 3))
			ExpectNext(it, 3, false)

			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 1, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{0, 1, 2, 3}))
		})
		It("add mixed order", func() {
			buf := NewOrderedBuffer(log)

			promise := Gather(buf)
			it := buf.Iterator()
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(3), 3))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(0), 0))
			ExpectNext(it, 0, false)

			buf.Add(NewEntry(Top(1), 1))
			ExpectNext(it, 1, false)

			buf.Add(NewEntry(Top(2), 2))
			ExpectNext(it, 2, true)
			ExpectNext(it, 3, false)

			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 1, 2, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{3, 0, 1, 2}))
		})

		It("add mixed order filtered", func() {
			buf := NewOrderedBuffer(log)

			promise := Gather(buf)
			it := buf.Iterator()
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(3), 3))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(0), 0))
			ExpectNext(it, 0, false)

			buf.Add(NewEntry(Top(2), 2))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Top(1), 1, false))
			ExpectNext(it, 2, true)
			ExpectNext(it, 3, false)

			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 2, 3}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{3, 0, 2, 1}))
		})

		It("exploded", func() {
			buf := NewOrderedBuffer(log)
			promise := Gather(buf)
			it := buf.Iterator()
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Index{0, 1}, 11, 2))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Index{0, 0}, 10, 2))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Index{0}, 0, SubEntries(2)))
			ExpectNext(it, 0, true)
			ExpectNext(it, 10, true)
			ExpectNext(it, 11, false)

			buf.Add(NewEntry(Index{1}, 1, SubEntries(1)))
			ExpectNext(it, 1, false)

			buf.Add(NewEntry(Index{2}, 2, SubEntries(1)))
			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Add(NewEntry(Index{1, 0}, 20, 1))
			ExpectNext(it, 20, true)
			ExpectNext(it, 2, false)

			buf.Add(NewEntry(Index{2, 0}, 30, 1))
			ExpectNext(it, 30, false)

			Expect(it.(CheckNext).CheckNext()).To(BeFalse())

			buf.Close()
			Expect(promise.Get()).To(Equal([]interface{}{0, 10, 11, 1, 20, 2, 30}))
			Expect(data.Slice(ValueIterable(buf))).To(Equal([]interface{}{11, 10, 0, 1, 2, 20, 30}))
		})
	})
})
