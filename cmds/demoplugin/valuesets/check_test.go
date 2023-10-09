// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package valuesets

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/cache"
)

const ARCH = "/tmp/ctf"
const COMP = "acme.org/test"
const VERS = "1.0.0"
const PROV = "acme.org"

var _ = Describe("demoplugin", func() {
	var env *testhelper.TestEnv

	BeforeEach(func() {
		env = testhelper.NewTestEnv()

		cache.DirectoryCache.Reset()
		plugindirattr.Set(env.OCMContext(), "testdata")

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERS, func() {
				env.Provider(PROV)
			})
		})
		env.RSAKeyPair(PROV)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("add check routing slip entry", func() {
		buf := bytes.NewBuffer(nil)
		MustBeSuccessful(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROV, "check", "--checkStatus", "test=passed", "--checkMessage", "test=25 tests successful"))
		Expect(buf.String()).To(Equal(""))

		buf.Reset()
		MustBeSuccessful(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, PROV))
		Expect(buf.String()).To(StringMatchTrimmedWithContext(`
COMPONENT-VERSION   NAME     TYPE  TIMESTAMP            DESCRIPTION
acme.org/test:1.0.0 acme.org check .{20} test: passed
`))
		buf.Reset()
		MustBeSuccessful(env.CatchOutput(buf).Execute("get", "routingslip", ARCH, PROV, "-oyaml"))
		Expect(buf.String()).To(StringMatchTrimmedWithContext(`message: 25 tests successful`))
	})
})
