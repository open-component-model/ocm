package errors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/errors"
)

var _ = Describe("errors", func() {
	Context("ErrReadOnly", func() {
		It("identifies kind error", func() {
			uerr := errors.ErrReadOnly("KIND", "obj")

			Expect(errors.IsErrReadOnlyKind(uerr, "KIND")).To(BeTrue())
			Expect(errors.IsErrReadOnlyKind(uerr, "other")).To(BeFalse())

		})
		It("message with elem", func() {
			uerr := errors.ErrReadOnly("KIND", "obj")

			Expect(uerr.Error()).To(Equal("KIND \"obj\" is readonly"))
		})
		It("message without elem", func() {
			uerr := errors.ErrReadOnly()

			Expect(uerr.Error()).To(Equal("readonly"))
		})
	})
	Context("ErrUnkown", func() {
		It("identifies kind error", func() {
			uerr := errors.ErrUnknown("KIND", "obj")

			Expect(errors.IsErrUnknownKind(uerr, "KIND")).To(BeTrue())
			Expect(errors.IsErrUnknownKind(uerr, "other")).To(BeFalse())

		})
		It("find error in history", func() {
			uerr := errors.ErrUnknown("KIND", "obj")
			werr := errors.Wrapf(uerr, "wrapped")

			Expect(errors.IsErrUnknownKind(werr, "KIND")).To(BeTrue())
			Expect(errors.IsErrUnknownKind(werr, "other")).To(BeFalse())
		})
	})

})
