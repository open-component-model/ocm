package data

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("slice", func() {
	Context("move", func() {
		It("backward move non-overlap", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(5, 7, 1)
			Expect(data).To(Equal(IndexedSliceAccess{0, 5, 6, 1, 2, 3, 4, 7, 8}))
		})
		It("backward move replace", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(3, 5, 1)
			Expect(data).To(Equal(IndexedSliceAccess{0, 3, 4, 1, 2, 5, 6, 7, 8}))
		})
		It("backward move in between", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(2, 4, 1)
			Expect(data).To(Equal(IndexedSliceAccess{0, 2, 3, 1, 4, 5, 6, 7, 8}))
		})
		It("forward move non-overlap", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(1, 3, 5)
			Expect(data).To(Equal(IndexedSliceAccess{0, 3, 4, 1, 2, 5, 6, 7, 8}))
		})
		It("forward move replace", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(1, 3, 3)
			Expect(data).To(Equal(IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}))
		})
		It("forward move in between", func() {
			data := IndexedSliceAccess{0, 1, 2, 3, 4, 5, 6, 7, 8}
			data.Move(1, 3, 4)
			Expect(data).To(Equal(IndexedSliceAccess{0, 3, 1, 2, 4, 5, 6, 7, 8}))
		})
	})
})
