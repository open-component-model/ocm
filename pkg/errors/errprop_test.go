package errors_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/errors"
)

func errfunc(succeed bool) func() error {
	if succeed {
		return func() error { return nil }
	}
	return func() error { return fmt.Errorf("error occurred") }
}

func testFunc(msg string, err error, succeed bool) (efferr error) {
	defer errors.PropagateErrorf(&efferr, errfunc(succeed), msg)
	return err
}

var _ = Describe("finalizer", func() {
	Context("without context", func() {
		It("succeeds", func() {
			Expect(testFunc("", nil, true)).To(Succeed())
		})

		It("fails ", func() {
			err := testFunc("", fmt.Errorf("failed"), true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed"))
		})

		It("succeeds with failing finalizer", func() {
			err := testFunc("", nil, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("error occurred"))
		})

		It("fails with failing finalizer", func() {
			err := testFunc("", fmt.Errorf("failed"), false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("{failed, error occurred}"))
		})
	})

	Context("with context", func() {
		It("succeeds", func() {
			Expect(testFunc("context", nil, true)).To(Succeed())
		})

		It("fails ", func() {
			err := testFunc("context", fmt.Errorf("failed"), true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("context: failed"))
		})

		It("succeeds with failing finalizer", func() {
			err := testFunc("context", nil, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("context: error occurred"))
		})

		It("fails with failing finalizer", func() {
			err := testFunc("context", fmt.Errorf("failed"), false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("context: {failed, error occurred}"))
		})
	})
})
