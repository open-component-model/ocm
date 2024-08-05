package runtimefinalizer_test

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/runtimefinalizer"
)

type ObjectType struct {
	kind string
	id   runtimefinalizer.ObjectIdentity
	fi   *runtimefinalizer.RuntimeFinalizer
}

func NewOType(kind string, r *runtimefinalizer.RuntimeFinalizationRecoder) *ObjectType {
	id := runtimefinalizer.NewObjectIdentity(kind)
	o := &ObjectType{
		kind: kind,
		id:   id,
		fi:   runtimefinalizer.NewRuntimeFinalizer(id, r),
	}
	return o
}

func (o *ObjectType) Id() runtimefinalizer.ObjectIdentity {
	return o.id
}

var _ = Describe("runtime finalizer", func() {
	It("finalize with arbitrary method", func() {
		r := &runtimefinalizer.RuntimeFinalizationRecoder{}

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
		Expect(r.Get()).To(Equal([]runtimefinalizer.ObjectIdentity{id1}))

		o2 = nil
		runtime.GC()
		time.Sleep(time.Second)
		Expect(r.Get()).To(Equal([]runtimefinalizer.ObjectIdentity{id1, id2}))
	})
})
