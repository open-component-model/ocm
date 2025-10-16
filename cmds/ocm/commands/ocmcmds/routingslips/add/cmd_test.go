package add_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip/types/comment"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	PROVIDER = "acme.org"
	OTHER    = "other.org"
)

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
