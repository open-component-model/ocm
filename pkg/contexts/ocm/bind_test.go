// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	me "github.com/open-component-model/ocm/pkg/contexts/ocm"
)

var _ = Describe("area test", func() {
	It("binds to Go context", func() {
		ctx := context.Background()

		mine := me.New()
		nctx := mine.BindTo(ctx)

		me.FromContext(nctx)
		Expect(me.FromContext(nctx)).To(BeIdenticalTo(mine))
	})
})
