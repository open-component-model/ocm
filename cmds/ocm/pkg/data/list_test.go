package data

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("linked list", func() {

	Context("add", func() {

		It("end", func() {
			list := NewLinkedList()

			list.Append(1).Append(2)

			Expect(Slice(list)).To(Equal([]interface{}{1, 2}))
		})

		It("start", func() {
			list := NewLinkedList()

			list.Append(1).Insert(2).Insert(3)

			Expect(Slice(list)).To(Equal([]interface{}{3, 2, 1}))
		})
	})
})
