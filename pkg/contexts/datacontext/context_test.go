// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package datacontext_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

var _ = Describe("area test", func() {
	It("can be garbage collected", func() {
		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.DebugLevel, me.Realm))

		ctx := me.New()
		Expect(ctx.IsIdenticalTo(ctx)).To(BeTrue())

		ctx2 := ctx.AttributesContext()
		Expect(ctx.IsIdenticalTo(ctx2)).To(BeTrue())

		ctx3 := me.New()
		Expect(ctx.IsIdenticalTo(ctx3)).To(BeFalse())
	})
})
