// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package refmgmt_test

import (
	"fmt"
	"io"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	refmgmt2 "github.com/open-component-model/ocm/pkg/refmgmt"
)

// Objectbase is the base interface for the
// object type to be wrapped.
type ObjectBase interface {
	io.Closer

	Value() (string, error)
}

////////////////////////////////////////////////////////////////////////////////

// Object is the final user facing interface.
// It includes the base interface plus the Dup method.
type Object interface {
	ObjectBase
	Dup() (Object, error)
}

////////////////////////////////////////////////////////////////////////////////

// object is the implementation type for the bse object.
type object struct {
	lock   sync.Mutex
	closed bool
	value  string
}

func (o *object) Value() (string, error) {
	if o.closed {
		return "", fmt.Errorf("should not happen")
	}
	return o.value, nil
}

func (o *object) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.closed {
		return refmgmt2.ErrClosed
	}
	o.closed = true
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// view is the view object used to wrap the base object.
// It forwards all methods to the base object using the
// Execute function of the manager, to assure execution
// on non-closed views, only.
type view struct {
	*refmgmt2.View[Object]
	obj ObjectBase
}

func (v *view) Value() (string, error) {
	value := ""

	err := v.Execute(func() (err error) {
		value, err = v.obj.Value() // forward to viewd object
		return
	})
	return value, err
}

// creator is the view object creator based on
// the base object and the view manager.
func creator(obj ObjectBase, v *refmgmt2.View[Object]) Object {
	return &view{v, obj}
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("view management wrapper", func() {
	It("wraps object", func() {
		o := &object{value: "test"}

		v := refmgmt2.WithView[ObjectBase, Object](o, creator)
		Expect(v.Value()).To(Equal("test"))

		d := Must(v.Dup())
		Expect(d.Value()).To(Equal("test"))

		MustBeSuccessful(d.Close())
		Expect(o.closed).To(BeFalse())
		ExpectError(d.Value()).To(Equal(refmgmt2.ErrClosed))
		Expect(v.Value()).To(Equal("test"))

		MustBeSuccessful(v.Close())
		Expect(o.closed).To(BeTrue())
		ExpectError(v.Value()).To(Equal(refmgmt2.ErrClosed))
	})
})
