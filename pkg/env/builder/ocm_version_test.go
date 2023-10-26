// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package builder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
)

const ARCH = "/tmp/ctf"
const ARCH2 = "/tmp/ctf2"
const PROVIDER = "mandelsoft"
const VERSION = "v1"
const COMPONENT = "github.com/mandelsoft/test"
const OUT = "/tmp/res"

var _ = Describe("Transfer handler", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
		compositionmodeattr.Set(env.OCMContext(), true)
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					TestDataResource(env)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("add ocm resource", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))

		Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))

		r := Must(cv.GetResourceByIndex(0))
		a := Must(r.Access())
		Expect(a.GetType()).To(Equal(localblob.Type))

		data := Must(ocmutils.GetResourceData(r))
		Expect(string(data)).To(Equal(S_TESTDATA))
	})
})
