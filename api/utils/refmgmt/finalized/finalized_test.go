package finalized_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/refmgmt"
	"ocm.software/ocm/api/utils/refmgmt/finalized"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

type Interface interface {
	GetId() runtimefinalizer.ObjectIdentity
	GetSelf() Interface

	GetRefId() runtimefinalizer.ObjectIdentity
}

type object struct {
	refmgmt.Allocatable
	recorder *runtimefinalizer.RuntimeFinalizationRecoder
	name     runtimefinalizer.ObjectIdentity
}

func (o *object) GetId() runtimefinalizer.ObjectIdentity {
	return o.name
}

func (o *object) GetSelf() Interface {
	v, _ := newView(o)
	return v
}

func (o *object) cleanup() error {
	o.recorder.Record(runtimefinalizer.ObjectIdentity(o.name))
	return nil
}

type view struct {
	ref *finalized.FinalizedRef
	*object
}

var _ Interface = (*view)(nil)

func newView(o *object) (Interface, error) {
	ref, err := finalized.NewFinalizedView(o.Allocatable, runtimefinalizer.NewObjectIdentity("test"), o.recorder)
	if err != nil {
		return nil, err
	}
	return &view{
		ref:    ref,
		object: o,
	}, nil
}

func (v *view) GetRefId() runtimefinalizer.ObjectIdentity {
	return v.ref.GetRefId()
}

func New(name string, rec *runtimefinalizer.RuntimeFinalizationRecoder) Interface {
	o := &object{
		name:     runtimefinalizer.ObjectIdentity(name),
		recorder: rec,
	}
	o.Allocatable = refmgmt.NewAllocatable(o.cleanup, true)

	v, _ := newView(o)
	return v
}

////////////////////////////////////////////////////////////////////////////////

var _ = Describe("finalized ref", func() {
	var rec *runtimefinalizer.RuntimeFinalizationRecoder

	BeforeEach(func() {
		rec = &runtimefinalizer.RuntimeFinalizationRecoder{}
	})

	It("cleanup ref and object", func() {
		o := New("test", rec)
		orefid := o.GetRefId()
		id := o.GetId()

		o = nil

		runtime.GC()
		time.Sleep(time.Second)

		Expect(rec.Get()).To(ConsistOf(
			id,
			orefid,
		))
	})

	It("cleanup ref after ref and then object", func() {
		o := New("test", rec)
		o2 := o
		orefoid := o.GetRefId()
		Expect(o2.GetRefId()).To(Equal(orefoid))

		id := o.GetId()
		r := o.GetSelf()
		rrefid := r.GetRefId()

		Expect(r.GetId()).To(Equal(id))
		Expect(orefoid).NotTo(Equal(rrefid))

		o.GetId()
		o = nil
		runtime.GC()
		time.Sleep(time.Second)

		Expect(rec.Get()).To(ConsistOf())

		r.GetId()
		r = nil
		runtime.GC()
		time.Sleep(time.Second)

		Expect(rec.Get()).To(ConsistOf(
			rrefid,
		))

		o2.GetId()
		o2 = nil
		runtime.GC()
		time.Sleep(time.Second)

		Expect(rec.Get()).To(ContainElements(
			orefoid,
			rrefid,
			id,
		))
	})
})
