package routingslips_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/extensions/labels/routingslip"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/plugin/registration"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH     = "/tmp/ca"
	VERSION  = "v1"
	COMP     = "test.de/x"
	PROVIDER = "acme.org"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv
	var plugins TempPluginDir

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
		plugins = Must(ConfigureTestPlugins(env, "testdata"))

		registry := plugincacheattr.Get(ctx)
		Expect(registration.RegisterExtensions(ctx)).To(Succeed())
		p := registry.Get("test")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		plugins.Cleanup()
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
