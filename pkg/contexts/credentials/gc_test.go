// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/finalizer"
)

var _ = Describe("area test", func() {
	It("can be garbage collected", func() {
		ctx := me.New()

		r := finalizer.GetRuntimeFinalizationRecorder(ctx)
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
