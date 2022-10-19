// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
)

type SpecHandler struct {
	name string
}

var _ core.RepositorySpecHandler = (*SpecHandler)(nil)

func (s SpecHandler) MapReference(ctx core.Context, u *core.UniformRepositorySpec) (core.RepositorySpec, error) {
	return nil, nil
}

var _ = Describe("spec handlers test", func() {
	var reg core.RepositorySpecHandlers

	BeforeEach(func() {
		reg = core.NewRepositorySpecHandlers()
	})

	It("copies registries", func() {
		mine := &SpecHandler{"mine"}

		reg.Register(mine, "arttype")

		h := reg.GetHandlers("arttype")
		Expect(h).To(Equal([]core.RepositorySpecHandler{mine}))

		copy := reg.Copy()
		new := &SpecHandler{"copy"}
		copy.Register(new, "arttype")

		h = reg.GetHandlers("arttype")
		Expect(h).To(Equal([]core.RepositorySpecHandler{mine}))

		h = copy.GetHandlers("arttype")
		Expect(h).To(Equal([]core.RepositorySpecHandler{mine, new}))
	})
})
