// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const COMPONENT2 = "github.com/mandelsoft/test2"
const COMPONENT3 = "github.com/mandelsoft/test3"
const OUT = "/tmp/res"
const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"
const SIGNATURE = "test"
const SIGN_ALGO = rsa.Algorithm

var _ = Describe("resolving local resource references", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(NewEnvironment())

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
			env.Component(COMPONENT2, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("ref", COMPONENT, VERSION)
					env.Resource("otherdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "otherdata")
					})
				})
			})

			env.Component(COMPONENT3, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("nested", COMPONENT2, VERSION)
					env.Resource("topdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "topdata")
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("resolves a direct local resource", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(src)
		cv := Must(src.LookupComponentVersion(COMPONENT3, VERSION))
		defer Close(cv)

		ref := metav1.NewResourceRef(metav1.NewIdentity("topdata"))

		res, eff, err := utils.ResolveResourceReference(cv, ref, nil)
		Expect(err).To(Succeed())
		m := Must(res.AccessMethod())
		data := Must(m.Get())
		Expect(string(data)).To(Equal("topdata"))
		MustBeSuccessful(eff.Close())
	})

	It("resolves an indirect resource", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(src)
		cv := Must(src.LookupComponentVersion(COMPONENT3, VERSION))
		defer Close(cv)

		ref := metav1.NewNestedResourceRef(metav1.NewIdentity("otherdata"), []metav1.Identity{metav1.NewIdentity("nested")})

		res, eff, err := utils.ResolveResourceReference(cv, ref, nil)
		Expect(err).To(Succeed())
		m := Must(res.AccessMethod())
		data := Must(m.Get())
		Expect(string(data)).To(Equal("otherdata"))
		MustBeSuccessful(eff.Close())
	})

	It("skips an intermediate component version", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(src)
		cv := Must(src.LookupComponentVersion(COMPONENT3, VERSION))
		defer Close(cv)

		ref := metav1.NewNestedResourceRef(metav1.NewIdentity("testdata"), []metav1.Identity{metav1.NewIdentity("nested"), metav1.NewIdentity("ref")})

		res, eff, err := utils.ResolveResourceReference(cv, ref, nil)
		Expect(err).To(Succeed())
		defer Close(eff)

		m := Must(res.AccessMethod())
		data := Must(m.Get())
		Expect(string(data)).To(Equal("testdata"))
	})
})
