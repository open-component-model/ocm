// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/config"
	local "github.com/open-component-model/ocm/pkg/contexts/credentials/core"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
)

var _ = Describe("builder test", func() {
	It("creates local", func() {
		ctx := local.Builder{}.New(datacontext.MODE_SHARED)

		Expect(ctx.AttributesContext()).To(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).To(BeIdenticalTo(local.DefaultRepositoryTypeScheme))

		Expect(ctx.ConfigContext()).To(BeIdenticalTo(config.DefaultContext()))
	})

	It("creates defaulted", func() {
		ctx := local.Builder{}.New(datacontext.MODE_DEFAULTED)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).To(BeIdenticalTo(local.DefaultRepositoryTypeScheme))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.ConfigContext().ConfigTypes()).To(BeIdenticalTo(config.DefaultContext().ConfigTypes()))
	})

	It("creates configured", func() {
		ctx := local.Builder{}.New(datacontext.MODE_CONFIGURED)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).NotTo(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(ctx.RepositoryTypes().KnownTypeNames()).To(Equal(local.DefaultRepositoryTypeScheme.KnownTypeNames()))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(ctx.ConfigContext().ConfigTypes()).NotTo(BeIdenticalTo(config.DefaultContext().ConfigTypes()))
		Expect(ctx.ConfigContext().ConfigTypes().KnownTypeNames()).To(Equal(config.DefaultContext().ConfigTypes().KnownTypeNames()))
	})

	It("creates iniial", func() {
		ctx := local.Builder{}.New(datacontext.MODE_INITIAL)

		Expect(ctx.AttributesContext()).NotTo(BeIdenticalTo(datacontext.DefaultContext))
		Expect(ctx).NotTo(BeIdenticalTo(local.DefaultContext))
		Expect(ctx.RepositoryTypes()).NotTo(BeIdenticalTo(local.DefaultRepositoryTypeScheme))
		Expect(len(ctx.RepositoryTypes().KnownTypeNames())).To(Equal(0))

		Expect(ctx.ConfigContext()).NotTo(BeIdenticalTo(config.DefaultContext()))
		Expect(len(ctx.ConfigContext().ConfigTypes().KnownTypeNames())).To(Equal(0))
	})

})
