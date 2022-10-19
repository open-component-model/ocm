// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package show_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const ARCH = "/tmp/ctf"
const PROVIDER = "mandelsoft"
const COMPONENT = "github.com/mandelsoft/test"
const V13 = "v1.3"
const V131 = "v1.3.1"
const V132 = "v1.3.2"
const V132x = "v1.3.2-beta.1"
const V14 = "v1.4"
const V2 = "v2.0"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(V13, func() {
					env.Provider(PROVIDER)
				})
				env.Version(V131, func() {
					env.Provider(PROVIDER)
				})
				env.Version(V132, func() {
					env.Provider(PROVIDER)
				})
				env.Version(V132x, func() {
					env.Provider(PROVIDER)
				})
				env.Version(V14, func() {
					env.Provider(PROVIDER)
				})
				env.Version(V2, func() {
					env.Provider(PROVIDER)
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists version", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("ocm", "versions", "show", "--repo", ARCH, COMPONENT)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
v1.3.2-beta.1
v1.3.2
v1.4
v2.0
`))
	})

	It("lists filtered versions", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("ocm", "versions", "show", "--repo", ARCH, COMPONENT, "1.3.x", "1.4")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3
v1.3.1
v1.3.2
v1.4
`))
	})

	It("lists filtered semantic versions", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("ocm", "versions", "show", "--semantic", "--repo", ARCH, COMPONENT, "1.3", "1.4")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
1.3.0
1.3.1
1.3.2
1.4.0
`))
	})

	It("lists filters latest", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("ocm", "versions", "show", "--latest", "--repo", ARCH, COMPONENT, "1.3")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(`
v1.3.2
`))
	})
})
