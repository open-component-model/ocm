// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/labels/routingslip/types/comment"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const PROVIDER = "acme.org"
const OTHER = "other.org"

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
		env.RSAKeyPair(PROVIDER, OTHER)
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("adds entry by generic entry option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, comment.Type, "--entry", "comment: first entry")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		Expect(slip.Len()).To(Equal(1))
		Expect(Must(slip.Get(0).Payload.Evaluate(env.OCMContext())).Describe(env.OCMContext())).To(Equal("Comment: first entry"))
	})

	It("adds entry by explicit field option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, comment.Type, "--comment", "first entry")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		Expect(slip.Len()).To(Equal(1))
		Expect(Must(slip.Get(0).Payload.Evaluate(env.OCMContext())).Describe(env.OCMContext())).To(Equal("Comment: first entry"))
	})

	It("adds dynamic entry by generic entry option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, "arbitrary", "--entry", "comment: first entry")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		Expect(slip.Len()).To(Equal(1))
		Expect(Must(slip.Get(0).Payload.Evaluate(env.OCMContext())).Describe(env.OCMContext())).To(Equal("comment: first entry"))
	})

	It("fails for dynamic entry with additional explicit option", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, "arbitrary", "--comment=test", "--entry", "comment: first entry")).To(MatchError(`unexpected options comment`))
	})

	DescribeTable("adds entry with slip link", func(args []string) {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, PROVIDER, "comment", "--comment", "first entry")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		Expect(env.CatchOutput(buf).Execute("add", "routingslip", ARCH, OTHER, "comment", "--comment", "other succeeded")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		Expect(env.CatchOutput(buf).Execute(append([]string{"add", "routingslip", ARCH, OTHER, "comment", "--comment", "link"}, args...)...)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
`))
		repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")
		cv := Must(repo.LookupComponentVersion(COMP, VERSION))
		defer Close(cv, "cv")
		slip := Must(routingslip.GetSlip(cv, PROVIDER))
		Expect(slip.Len()).To(Equal(1))
		Expect(Must(slip.Get(0).Payload.Evaluate(env.OCMContext())).Describe(env.OCMContext())).To(Equal("Comment: first entry"))

		slip2 := Must(routingslip.GetSlip(cv, OTHER))
		Expect(slip2.Len()).To(Equal(2))
		Expect(slip2.Get(1).Links).To(Equal([]routingslip.Link{{Name: PROVIDER, Digest: slip.Get(0).Digest}}))
	},
		Entry("for slip", []string{"--links=" + PROVIDER}),
		Entry("for all slips", []string{"--links=all"}),
	)
})
