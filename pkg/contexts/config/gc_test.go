package config_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/pkg/contexts/config"
	"github.com/open-component-model/ocm/pkg/runtimefinalizer"
)

var _ = Describe("area test", func() {
	It("can be garbage collected", func() {
		ctx := me.New()

		r := runtimefinalizer.GetRuntimeFinalizationRecorder(ctx)
		Expect(r).NotTo(BeNil())

		runtime.GC()
		time.Sleep(time.Second)
		ctx.GetType()
		Expect(r.Get()).To(BeNil())

		ctx = nil
		for i := 0; i < 100; i++ {
			runtime.GC()
			time.Sleep(time.Millisecond)
		}

		Expect(r.Get()).To(ContainElement(ContainSubstring(me.CONTEXT_TYPE)))
	})
})
