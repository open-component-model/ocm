// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package exception

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/errors"
)

var dump = fmt.Errorf("dump")

func callee(e error) (int, error) {
	if e == dump {
		a := 0
		_ = 1 / a
	}
	return 0, e
}

func _caller(e error, args ...interface{}) {
	if len(args) == 0 {
		Must1(callee(e))
	} else {
		Must1f(R1(callee(e)), args[0].(string), args[1:]...)
	}
}

func caller(e error, args ...interface{}) (err error) {
	defer PropagateException(&err)
	_caller(e, args...)
	return nil
}

var _ = Describe("exceptions", func() {
	It("succeeds", func() {
		Expect(caller(nil)).To(BeNil())
	})

	It("propagates", func() {
		err := fmt.Errorf("test error")
		Expect(caller(err)).To(Equal(err))
	})

	It("dumps", func() {
		defer func() {
			r := recover()
			Expect(r).NotTo(BeNil())
			Expect(fmt.Sprintf("%s", r)).To(Equal("runtime error: integer divide by zero"))
		}()
		caller(dump)
	})

	It("propagates with context", func() {
		err := fmt.Errorf("test error")

		prop := caller(err, "test")
		Expect(prop).NotTo(BeNil())
		Expect(errors.Unwrap(prop)).To(Equal(err))
		Expect(prop.Error()).To(Equal("test: test error"))
	})

	It("propagates with outer context", func() {
		caller := func(e error, args ...interface{}) (err error) {
			defer PropagateExceptionf(&err, "outer")
			_caller(e, args...)
			return nil
		}

		err := fmt.Errorf("test error")

		prop := caller(err)
		Expect(prop).NotTo(BeNil())
		Expect(errors.Unwrap(prop)).To(Equal(err))
		Expect(prop.Error()).To(Equal("outer: test error"))

		prop = caller(err, "test")
		Expect(prop).NotTo(BeNil())
		Expect(errors.Unwrap(errors.Unwrap(prop))).To(Equal(err))
		Expect(prop.Error()).To(Equal("outer: test: test error"))
	})
})
