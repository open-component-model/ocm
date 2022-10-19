// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
)

var _ = Describe("helper", func() {

	It("should inject a new repository context if none is defined", func() {
		cd := &compdesc.ComponentDescriptor{}
		compdesc.DefaultComponent(cd)

		repoCtx := ocireg.NewRepositorySpec("example.com", nil)
		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		Expect(cd.AddRepositoryContext(repoCtx)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(1))

		repoCtx2 := ocireg.NewRepositorySpec("example.com/dev", nil)
		Expect(cd.AddRepositoryContext(repoCtx2)).To(Succeed())
		Expect(cd.RepositoryContexts).To(HaveLen(2))
	})

})
