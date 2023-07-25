// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	local "github.com/open-component-model/ocm/v2/pkg/contexts/config/internal"
	"github.com/open-component-model/ocm/v2/pkg/contexts/datacontext"
)

var _ = Describe("builder test", func() {
	It("creates local", func() {
		ctx := local.Builder{}.New(datacontext.MODE_SHARED)

		Expect(ctx.AttributesContext()).To(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.ConfigTypes()).To(BeIdenticalTo(local.DefaultConfigTypeScheme))
	})

	It("creates configured", func() {
		ctx := local.Builder{}.New(datacontext.MODE_CONFIGURED)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.ConfigTypes()).NotTo(BeIdenticalTo(local.DefaultConfigTypeScheme))
		Expect(ctx.ConfigTypes().KnownTypeNames()).To(Equal(local.DefaultConfigTypeScheme.KnownTypeNames()))
	})

	It("creates iniial", func() {
		ctx := local.Builder{}.New(datacontext.MODE_INITIAL)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.ConfigTypes()).NotTo(BeIdenticalTo(local.DefaultConfigTypeScheme))
		Expect(len(ctx.ConfigTypes().KnownTypeNames())).To(Equal(0))
	})

})
