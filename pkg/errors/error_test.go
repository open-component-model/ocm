// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/errors"
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
