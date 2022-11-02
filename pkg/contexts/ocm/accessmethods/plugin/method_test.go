// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
)

const ARCH = "ctf"
const COMP = "github.com/mandelsoft/comp"
const VERS = "1.0.0"
const PROVIDER = "mandelsoft"

var _ = Describe("setup plugin cache", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Builder

	var accessSpec ocm.AccessSpec

	BeforeEach(func() {
		var err error

		accessSpec, err = ocm.NewGenericAccessSpec(`
type: test
someattr: value
`)
		Expect(err).To(Succeed())

		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
		Expect(registry.RegisterExtensions()).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("registers access methods", func() {

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", VERS, "PlainText", metav1.ExternalRelation, func() {
						env.Access(accessSpec)
					})
				})
			})
		})

		repo, err := ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env)
		Expect(err).To(Succeed())
		defer Close(repo)

		cv, err := repo.LookupComponentVersion(COMP, VERS)
		Expect(err).To(Succeed())
		defer Close(cv)

		r, err := cv.GetResourceByIndex(0)
		Expect(err).To(Succeed())

		m, err := r.AccessMethod()
		Expect(err).To(Succeed())
		Expect(m.MimeType()).To(Equal("plain/text"))

		data, err := m.Get()
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("test content\n"))
	})
})
