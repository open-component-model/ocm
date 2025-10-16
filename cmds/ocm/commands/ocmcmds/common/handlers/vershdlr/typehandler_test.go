package vershdlr_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	ARCH  = "ctf"
	COMP  = "acme.org/comp1"
	VERS1 = "1.0.0"
	VERS2 = "2.0.0"
)

var _ = Describe("version handler", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERS1)
				env.Version(VERS2)
			})
		})

		env.OCMContext().AddResolverRule(COMP, Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, env)))
	})

	AfterEach(func() {
		vfs.Cleanup(env)
	})

	Context("using resolvers", func() {
		It("resolves versions", func() {
			var buf bytes.Buffer
			MustBeSuccessful(env.CatchOutput(&buf).Execute("list", "cv", COMP))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
COMPONENT      VERSION MESSAGE
acme.org/comp1 1.0.0   
acme.org/comp1 2.0.0   
`))
		})

		It("provides error for non-matching resolver", func() {
			var buf bytes.Buffer
			MustBeSuccessful(env.CatchOutput(&buf).Execute("list", "cv", "acme.org/dummy"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
COMPONENT      VERSION MESSAGE
acme.org/dummy         <unknown component version>
`))
		})

		It("provides error for non-matching component", func() {
			var buf bytes.Buffer
			MustBeSuccessful(env.CatchOutput(&buf).Execute("list", "cv", COMP+"/"+"dummy"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
COMPONENT            VERSION MESSAGE
acme.org/comp1/dummy         <unknown component version>
`))
		})
	})
})
