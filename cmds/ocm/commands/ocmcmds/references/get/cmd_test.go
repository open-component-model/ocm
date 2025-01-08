package get_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/utils/accessio"
)

const (
	CA       = "/tmp/ca"
	CTF      = "/tmp/ctf"
	VERSION  = "v1"
	COMP     = "test.de/x"
	COMP2    = "test.de/y"
	COMP3    = "test.de/z"
	PROVIDER = "mandelsoft"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("lists single reference in component archive", func() {
		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Reference("test", COMP2, VERSION)
			env.Reference("withid", COMP3, VERSION, func() {
				env.ExtraIdentity("id", "test")
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "references", "-o", "wide", CA)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
NAME   COMPONENT VERSION IDENTITY
test   test.de/y v1      "name"="test"
withid test.de/z v1      "id"="test","name"="withid"
`))
	})

	Context("with closure", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP2, VERSION, func() {
					env.Provider(PROVIDER)
					env.Reference("withid", COMP3, VERSION, func() {
						env.ExtraIdentity("id", "test")
					})
				})
				env.ComponentVersion(COMP3, VERSION, func() {
					env.Provider(PROVIDER)
				})
			})
			env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERSION, func() {
				env.Provider(PROVIDER)
				env.Reference("test", COMP2, VERSION)
			})
		})
		It("lists single reference in component archive", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "--lookup", CTF, "-r", "-o", "wide", CA)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REFERENCEPATH              NAME   COMPONENT VERSION IDENTITY
test.de/x:v1               test   test.de/y v1      "name"="test"
test.de/x:v1->test.de/y:v1 withid test.de/z v1      "id"="test","name"="withid"
`))
		})
		It("lists flat tree in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "-o", "tree", "--lookup", CTF, CA)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENTVERSION NAME COMPONENT VERSION IDENTITY
└─ test.de/x:v1                         
   └─            test test.de/y v1      "name"="test"
`))
		})

		It("list reference closure in ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "references", "-r", "-o", "tree", "--lookup", CTF, CA)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENTVERSION NAME   COMPONENT VERSION IDENTITY
└─ test.de/x:v1                           
   └─ ⊗          test   test.de/y v1      "name"="test"
      └─         withid test.de/z v1      "id"="test","name"="withid"
`))
		})
	})
})
