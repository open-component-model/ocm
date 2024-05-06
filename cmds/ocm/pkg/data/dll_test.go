package data

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("simple data processing", func() {

	Context("add", func() {

		It("end", func() {
			root := DLL{}

			d1 := NewDLL(1)

			root.Append(d1)
			Expect(root.Next()).To(BeIdenticalTo(d1))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeIdenticalTo(&root))

			Expect(Slice(&DLLRoot{root: root})).To(Equal([]interface{}{d1}))
		})
		It("middle", func() {
			root := DLL{}

			d1 := NewDLL(1)
			d2 := NewDLL(2)

			root.Append(d2)
			root.Append(d1)

			Expect(root.Next()).To(BeIdenticalTo(d1))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeIdenticalTo(d2))
			Expect(d1.Prev()).To(BeIdenticalTo(&root))

			Expect(d2.Next()).To(BeNil())
			Expect(d2.Prev()).To(BeIdenticalTo(d1))

			Expect(Slice(&DLLRoot{root: root})).To(Equal([]interface{}{d1, d2}))
		})
	})
	Context("remove", func() {

		It("end", func() {
			root := DLL{}

			d1 := NewDLL(1)

			root.Append(d1)

			d1.Remove()

			Expect(root.Next()).To(BeNil())
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeNil())
		})
		It("middle", func() {
			root := DLL{}

			d1 := NewDLL(1)
			d2 := NewDLL(2)

			root.Append(d2)
			root.Append(d1)

			d1.Remove()

			Expect(root.Next()).To(BeIdenticalTo(d2))
			Expect(root.Prev()).To(BeNil())

			Expect(d1.Next()).To(BeNil())
			Expect(d1.Prev()).To(BeNil())

			Expect(d2.Next()).To(BeNil())
			Expect(d2.Prev()).To(BeIdenticalTo(&root))

		})
	})

})
