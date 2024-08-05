package datacontext_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/general"

	me "ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/utils/runtimefinalizer"
)

var _ = Describe("area test", func() {
	It("can be garbage collected", func() {
		ctx := me.New()
		r := runtimefinalizer.GetRuntimeFinalizationRecorder(ctx)
		id := ctx.GetId()
		Expect(me.GetContextRefCount(ctx)).To(Equal(1))
		ctx = nil
		runtime.GC()
		time.Sleep(time.Second)
		Expect(r.Get()).To(ConsistOf(id))
	})

	It("provides second reference", func() {
		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.DebugLevel, me.Realm))
		multiRefs := general.Conditional(me.MULTI_REF, 2, 1)

		ctx := me.New()
		Expect(me.GetContextRefCount(ctx)).To(Equal(1))

		actx := ctx.AttributesContext()
		Expect(me.GetContextRefCount(ctx)).To(Equal(multiRefs))

		r := runtimefinalizer.GetRuntimeFinalizationRecorder(ctx)
		Expect(r).NotTo(BeNil())

		runtime.GC()
		time.Sleep(time.Second)
		ctx.GetType()
		Expect(r.Get()).To(BeNil())

		actx.GetType()
		actx = nil
		runtime.GC()
		time.Sleep(time.Second)
		ctx.GetType()
		Expect(r.Get()).To(BeNil())
		Expect(me.GetContextRefCount(ctx)).To(Equal(1))

		ctx = nil
		for i := 0; i < 100; i++ {
			runtime.GC()
			time.Sleep(time.Millisecond)
		}

		Expect(r.Get()).To(ContainElement(ContainSubstring(me.CONTEXT_TYPE)))
	})

	It("creates views", func() {
		ctx := me.New()
		r := runtimefinalizer.GetRuntimeFinalizationRecorder(ctx)

		Expect(me.GetContextRefCount(ctx)).To(Equal(1))
		Expect(me.IsPersistentContextRef(ctx)).To(BeTrue())

		view := me.PersistentContextRef(ctx)
		Expect(me.GetContextRefCount(view)).To(Equal(1)) // reuse persistent ref
		Expect(me.IsPersistentContextRef(view)).To(BeTrue())

		non := view.AttributesContext()
		Expect(me.IsPersistentContextRef(non)).To(BeFalse())

		view2 := me.PersistentContextRef(non)
		Expect(me.GetContextRefCount(view2)).To(Equal(2)) // create new view
		Expect(me.IsPersistentContextRef(view2)).To(BeTrue())

		Expect(ctx.IsIdenticalTo(view)).To(BeTrue())
		Expect(ctx.IsIdenticalTo(view2)).To(BeTrue())

		ctx = nil
		view = nil
		view2 = nil

		runtime.GC()
		time.Sleep(time.Second)
		Expect(len(r.Get())).To(Equal(1)) // ref non is not persistent
	})
})
