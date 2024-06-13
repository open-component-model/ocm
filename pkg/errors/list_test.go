package errors

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("error list", func() {
	Context("without message", func() {
		It("handles no error", func() {
			Expect(ErrListf("").Result()).To(Succeed())
		})

		It("handles one error", func() {
			err := ErrListf("").Add(fmt.Errorf("e1")).Result()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("e1"))
		})

		It("handles two error2", func() {
			err := ErrListf("").Add(fmt.Errorf("e1"), fmt.Errorf("e2")).Result()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("{e1, e2}"))
		})
	})

	Context("with message", func() {
		It("handles no error", func() {
			Expect(ErrListf("msg").Result()).To(Succeed())
		})

		It("handles one error", func() {
			err := ErrListf("msg").Add(fmt.Errorf("e1")).Result()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("msg: e1"))
		})

		It("handles two error2", func() {
			err := ErrListf("msg").Add(fmt.Errorf("e1"), fmt.Errorf("e2")).Result()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("msg: {e1, e2}"))
		})
	})

	Context("is a", func() {
		It("handles single nested", func() {
			err := ErrListf("msg").Add(ErrInvalid()).Result()
			Expect(err).To(HaveOccurred())
			Expect(IsA(err, ErrInvalid())).To(BeTrue())
		})

		It("handles nested single nested", func() {
			err := ErrListf("msg").Add(ErrInvalid()).Result()
			Expect(err).To(HaveOccurred())
			Expect(IsA(err, ErrInvalid())).To(BeTrue())
		})

		It("gets single nested", func() {
			err := Wrapf(ErrListf("msg").Add(ErrInvalid("test")).Result(), "top")
			exp := &InvalidError{}
			Expect(err).To(HaveOccurred())
			Expect(As(err, &exp)).To(BeTrue())
			Expect(exp.Error()).To(Equal("\"test\" is invalid"))
		})
	})
})
