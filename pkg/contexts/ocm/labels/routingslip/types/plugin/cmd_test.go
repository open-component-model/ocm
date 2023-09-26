// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const PROVIDER = "acme.org"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
				})
			})
		})
		env.RSAKeyPair(PROVIDER)

		ctx := env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry := plugincacheattr.Get(ctx)
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds entry by plugin option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, "test", "--accessPath", "some path", "--mediaType", "media type")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		Expect(slip.Len()).To(Equal(1))
		Expect(Must(slip.Get(0).Payload.Evaluate(env.OCMContext())).Describe(env.OCMContext())).To(Equal("a test"))
	})
})
