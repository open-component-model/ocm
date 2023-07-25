// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugindirattr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/config"
	"github.com/open-component-model/ocm/v2/pkg/contexts/datacontext"
	me "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugindirattr"
)

var _ = Describe("attribute", func() {
	var ctx config.Context

	attr := "___test___"

	BeforeEach(func() {
		ctx = config.WithSharedAttributes(datacontext.New(nil)).New()
	})

	It("local setting", func() {
		Expect(me.Get(ctx)).NotTo(Equal(attr))
		Expect(me.Set(ctx, attr)).To(Succeed())
		Expect(me.Get(ctx)).To(BeIdenticalTo(attr))
	})
})
