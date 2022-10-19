// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/utils"
)

type Order []string

func F(n string, order *Order) func() error {
	return func() error {
		*order = append(*order, n)
		return nil
	}
}

var _ = Describe("finalizer", func() {
	It("finalize in revered order", func() {
		var finalize utils.Finalizer
		var order Order

		finalize.With(F("A", &order))
		finalize.With(F("B", &order))

		finalize.Finalize()

		Expect(order).To(Equal(Order{"B", "A"}))
		Expect(finalize.Length()).To(Equal(0))
	})

	It("is reusable after calling Finalize", func() {
		var finalize utils.Finalizer
		var order Order

		finalize.With(F("A", &order))
		finalize.With(F("B", &order))

		finalize.Finalize()
		order = nil

		finalize.With(F("C", &order))
		finalize.With(F("D", &order))

		finalize.Finalize()

		Expect(order).To(Equal(Order{"D", "C"}))
		Expect(finalize.Length()).To(Equal(0))
	})
})
