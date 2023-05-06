// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package finalizer_test

import (
	"fmt"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/exception"
	"github.com/open-component-model/ocm/pkg/finalizer"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

type Order []string

func F(n string, order *Order) func() error {
	return func() error {
		return A(n, order)
	}
}

func A(n string, order *Order) error {
	*order = append(*order, n)
	return nil
}

type closer struct {
	io.ReadCloser
	name  string
	order *Order
}

func Closer(n string, order *Order) io.ReadCloser {
	return &closer{nil, n, order}
}

func (c *closer) Close() error {
	return A(c.name, c.order)
}

var _ = Describe("finalizer", func() {
	It("finalize with arbitrary method", func() {
		var finalize finalizer.Finalizer
		var order Order

		finalize.With(finalizer.Calling2(A, "A", &order))
		Expect(order).To(BeNil())

		finalize.Finalize()

		Expect(order).To(Equal(Order{"A"}))
		Expect(finalize.Length()).To(Equal(0))
	})

	It("finalize in reversed order", func() {
		var finalize finalizer.Finalizer
		var order Order

		finalize.With(F("A", &order))
		finalize.With(F("B", &order))
		Expect(order).To(BeNil())

		finalize.Finalize()

		Expect(order).To(Equal(Order{"B", "A"}))
		Expect(finalize.Length()).To(Equal(0))
	})

	It("is reusable after calling Finalize", func() {
		var finalize finalizer.Finalizer
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

	It("separately finalizes a Nested finalizer", func() {
		var finalize finalizer.Finalizer
		var order Order

		finalize.With(F("A", &order))
		finalize.With(F("B", &order))

		{
			finalize := finalize.Nested()
			finalize.With(F("C", &order))
			finalize.Finalize()
			Expect(order).To(Equal(Order{"C"}))
		}

		{
			finalize := finalize.Nested()
			finalize.With(F("D", &order))
			finalize.Finalize()
			Expect(order).To(Equal(Order{"C", "D"}))
		}

		{
			finalize := finalize.Nested()
			finalize.With(F("E", &order))
		}

		finalize.Finalize()
		Expect(order).To(Equal(Order{"C", "D", "E", "B", "A"}))
		Expect(finalize.Length()).To(Equal(0))
	})

	It("separately finalizes new finalizers", func() {
		var finalize finalizer.Finalizer
		var order Order

		finalize.With(F("A", &order))
		finalize.With(F("B", &order))

		{
			finalize := finalize.New()
			finalize.With(F("C", &order))
		}

		{
			finalize := finalize.Nested()
			finalize.With(F("D", &order))
			finalize.Finalize()
			Expect(order).To(Equal(Order{"D"}))
		}

		{
			finalize := finalize.New()
			finalize.With(F("E", &order))
		}

		finalize.Finalize()
		Expect(order).To(Equal(Order{"D", "E", "C", "B", "A"}))
		Expect(finalize.Length()).To(Equal(0))
	})

	Context("with error propagation", func() {
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

	Context("with exceptions", func() {
		callee := func() {
			exception.Throw(fmt.Errorf("exception"))
		}
		caller := func() (err error) {
			var finalize finalizer.Finalizer

			defer finalize.CatchException().FinalizeWithErrorPropagation(&err)
			callee()
			return nil
		}

		It("catches exception from exception package", func() {
			err := caller()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("exception"))
		})
	})

	Context("transfering", func() {
		It("transfers actions", func() {
			var f finalizer.Finalizer
			var order Order

			f.With(F("first", &order))
			c := Closer("closer", &order)

			b := f.BindToReader(c, "bound")

			f.With(F("second", &order))

			MustBeSuccessful(f.Finalize())
			MustBeSuccessful(b.Close())

			Expect(order).To(Equal(Order{"second", "closer", "first"}))
		})

		It("transfers nested actions", func() {
			var f finalizer.Finalizer
			var order Order

			f.With(F("first", &order))
			n := f.Nested()
			n.With(F("nested", &order))

			c := Closer("closer", &order)

			b := n.BindToReader(c, "bound")
			n.With(F("next", &order))

			f.With(F("second", &order))

			MustBeSuccessful(f.Finalize())
			MustBeSuccessful(b.Close())

			Expect(order).To(Equal(Order{"second", "next", "first", "closer", "nested"}))
		})
	})
})

func errfunc(succeed bool) func() error {
	if succeed {
		return func() error { return nil }
	}
	return func() error { return fmt.Errorf("error occurred") }
}

func testFunc(msg string, err error, succeed bool) (efferr error) {
	var finalize finalizer.Finalizer

	defer finalize.FinalizeWithErrorPropagationf(&efferr, msg)
	finalize.With(errfunc(succeed))
	return err
}
