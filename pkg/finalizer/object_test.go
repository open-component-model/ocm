// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package finalizer_test

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/finalizer"
)

type ObjectType struct {
	kind string
	id   finalizer.ObjectIdentity
	fi   *finalizer.RuntimeFinalizer
}

func NewOType(kind string, r *finalizer.RuntimeFinalizationRecoder) *ObjectType {
	id := finalizer.NewObjectIdentity(kind)
	o := &ObjectType{
		kind: kind,
		id:   id,
		fi:   finalizer.NewRuntimeFinalizer(id, r),
	}
	return o
}

func (o *ObjectType) Id() finalizer.ObjectIdentity {
	return o.id
}

var _ = Describe("runtime finalizer", func() {
	It("finalize with arbitrary method", func() {
		r := &finalizer.RuntimeFinalizationRecoder{}

		o1 := NewOType("test1", r)
		o2 := NewOType("test1", r)

		id1 := o1.Id()
		id2 := o2.Id()

		runtime.GC()
		time.Sleep(time.Second)

		fmt.Printf("still used (%s,%s)\n", o1.Id(), o2.Id())
		Expect(len(r.Get())).To(Equal(0))

		o1 = nil
		runtime.GC()
		time.Sleep(time.Second)
		fmt.Printf("still used (%s)\n", o2.Id())
		Expect(r.Get()).To(Equal([]finalizer.ObjectIdentity{id1}))

		o2 = nil
		runtime.GC()
		time.Sleep(time.Second)
		Expect(r.Get()).To(Equal([]finalizer.ObjectIdentity{id1, id2}))
	})
})
