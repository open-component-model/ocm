// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/env/builder"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/open-component-model/ocm/v2/pkg/common/accessio"
	"github.com/open-component-model/ocm/v2/pkg/common/accessobj"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/attrs/plugindirattr"
	metav1 "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/plugin/plugins"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/ctf"
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
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
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

		repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo)

		cv := Must(repo.LookupComponentVersion(COMP, VERS))
		defer Close(cv)

		r := Must(cv.GetResourceByIndex(0))

		m := Must(r.AccessMethod())
		Expect(m.MimeType()).To(Equal("plain/text"))

		data := Must(m.Get())
		Expect(string(data)).To(Equal("test content\n{\"someattr\":\"value\",\"type\":\"test\"}\n"))
	})
})
