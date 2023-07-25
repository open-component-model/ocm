// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/internal"
)

type SpecHandler struct {
	name string
}

var _ internal.RepositorySpecHandler = (*SpecHandler)(nil)

func (s SpecHandler) MapReference(ctx internal.Context, u *internal.UniformRepositorySpec) (internal.RepositorySpec, error) {
	return nil, nil
}

var _ = Describe("spec handlers test", func() {
	var reg internal.RepositorySpecHandlers

	BeforeEach(func() {
		reg = internal.NewRepositorySpecHandlers()
	})

	It("copies registries", func() {
		mine := &SpecHandler{"mine"}

		reg.Register(mine, "arttype")

		h := reg.GetHandlers("arttype")
		Expect(h).To(Equal([]internal.RepositorySpecHandler{mine}))

		copy := reg.Copy()
		new := &SpecHandler{"copy"}
		copy.Register(new, "arttype")

		h = reg.GetHandlers("arttype")
		Expect(h).To(Equal([]internal.RepositorySpecHandler{mine}))

		h = copy.GetHandlers("arttype")
		Expect(h).To(Equal([]internal.RepositorySpecHandler{mine, new}))
	})
})
