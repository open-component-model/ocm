// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials_test

import (
	"runtime"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/v2/pkg/finalizer"
)

var _ = Describe("area test", func() {
	It("can access the default context", func() {
		ctx := credentials.New()

		r := finalizer.GetRuntimeFinalizationRecorder(ctx)
		Expect(r).NotTo(BeNil())

		runtime.GC()
		time.Sleep(time.Second)
		ctx.GetType()
		Expect(r.Get()).To(BeNil())

		ctx = nil
		runtime.GC()
		time.Sleep(time.Second)

		Expect(r.Get()).To(ContainElement(ContainSubstring(credentials.CONTEXT_TYPE)))
	})
})
