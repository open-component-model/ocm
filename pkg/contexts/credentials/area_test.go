// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package credentials_test

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
)

var DefaultContext = credentials.New()

var _ = Describe("area test", func() {
	It("can access the default context", func() {
		ctx := credentials.ForContext(context.TODO())
		Expect(ctx).NotTo(BeNil())
		Expect(reflect.TypeOf(ctx).String()).To(Equal("*core._context"))
	})
	It("can access the set context", func() {
		ctx := DefaultContext.BindTo(context.TODO())
		dctx := credentials.ForContext(ctx)
		Expect(dctx).NotTo(BeNil())
		Expect(reflect.TypeOf(dctx).String()).To(Equal("*core._context"))
		Expect(dctx).To(BeIdenticalTo(DefaultContext))
	})

})
