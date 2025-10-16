package list_test

import (
	"bytes"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/utils/accessio"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH      = "/tmp/ca"
	ARCH2     = "/tmp/ca2"
	VERSION   = "v1"
	VERSION11 = "v1.1"
	VERSION2  = "v2"
	COMP      = "test.de/x"
	COMP2     = "test.de/y"
	COMP3     = "test.de/z"
	PROVIDER  = "mandelsoft"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("list component archive", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("list", "components", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION MESSAGE
test.de/x v1      
`))
	})

	It("list component archive with refs", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Reference("ref", COMP2, VERSION)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("list", "components", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION MESSAGE
test.de/x v1                        
`))
	})

	It("lists ctf file", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMP, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
				})
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("list", "components", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION MESSAGE
test.de/x v1      
`))
	})

	Context("ctf with multiple versions", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
					})
				})
				env.Component(COMP, func() {
					env.Version(VERSION11, func() {
						env.Provider(PROVIDER)
					})
				})
				env.Component(COMP, func() {
					env.Version(VERSION2, func() {
						env.Provider(PROVIDER)
					})
				})
			})
		})

		It("lists all versions", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/x v1      
test.de/x v1.1    
test.de/x v2      
`))
		})

		It("lists latest version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--latest", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/x v2      
`))
		})

		It("lists constrained version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--constraints", ">1.0", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/x v1.1    
test.de/x v2      
`))
		})

		It("lists constrained version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--constraints", "1.x.x", "--latest", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/x v1.1    
`))
		})
	})

	Context("ctf", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMP2, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Reference("xx", COMP, VERSION)
					})
				})
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
					})
				})
			})
		})

		It("lists all components", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--repo", ARCH)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/x v1      
test.de/y v1      
`))
		})

		It("lists all components as json", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--repo", ARCH, "-o", "json")).To(Succeed())
			fmt.Printf("%s\n", buf.String())
			Expect(buf.String()).To(YAMLEqual(`
items:
- component: test.de/x
  version: v1
- component: test.de/y
  version: v1
`))
		})

		It("reports unknown in repo", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", "--repo", ARCH, COMP3+":v1")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/z v1      <unknown component version>
`))
		})

		It("reports unknown", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", ARCH+"//"+COMP3+":v1")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION MESSAGE
test.de/z v1      <unknown component version>
`))
		})

		It("reports unknown as json", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("list", "components", ARCH+"//"+COMP3+":v1", "-o", "json")).To(Succeed())
			Expect(buf.String()).To(YAMLEqual(`
items:
- component: test.de/z
  version: v1
  error: <unknown component version>
`))
		})
	})
})
