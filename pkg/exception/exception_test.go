package exception_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/errors"
	"github.com/open-component-model/ocm/pkg/exception"
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
		exception.Must1(callee(e))
	} else {
		exception.Must1f(exception.R1(callee(e)), args[0].(string), args[1:]...)
	}
}

func caller(e error, args ...interface{}) (err error) {
	defer exception.PropagateException(&err)
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
			defer exception.PropagateExceptionf(&err, "outer")
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

	Context("with matchers", func() {

		caller := func(e error, args ...interface{}) (err error) {
			defer exception.PropagateException(&err, exception.ByPrototypes(MyException{}))
			_caller(e, args...)
			return nil
		}

		It("catches matched exception", func() {
			err := caller(MyException{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("MyException"))
		})

		It("passes unmatched exception", func() {
			defer func() {
				r := recover()
				Expect(r).NotTo(BeNil())
				Expect(fmt.Sprintf("%s", r)).To(Equal("test"))
			}()
			err := caller(fmt.Errorf("test"))
			Expect(err).To(Succeed())
		})
	})
})

type MyException struct{}

func (_ MyException) Error() string {
	return "MyException"
}
