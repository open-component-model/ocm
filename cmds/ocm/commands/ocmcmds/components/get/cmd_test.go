package get_test

import (
	"bytes"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	compdescv3 "ocm.software/ocm/api/ocm/compdesc/versions/ocm.software/v3alpha1"
	"ocm.software/ocm/api/utils/accessio"
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

	It("get component archive", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION PROVIDER   REPOSITORY
test.de/x v1      mandelsoft /tmp/ca
`))
	})

	It("get component archive with refs", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Reference("ref", COMP2, VERSION)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "components", ARCH, "-r")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
REFERENCEPATH COMPONENT VERSION PROVIDER                    IDENTITY
              test.de/x v1      mandelsoft                  
test.de/x:v1  test.de/y v1      <unknown component version> "name"="ref"
`))
	})

	It("get component archive with refs as tree", func() {
		env.ComponentArchive(ARCH, accessio.FormatDirectory, COMP, VERSION, func() {
			env.Provider(PROVIDER)
			env.Reference("ref", COMP2, VERSION)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "components", ARCH, "-r", "-o", "tree")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
NESTING COMPONENT VERSION PROVIDER                    IDENTITY
└─ ⊗    test.de/x v1      mandelsoft                  
   └─   test.de/y v1      <unknown component version> "name"="ref"
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
		Expect(env.CatchOutput(buf).Execute("get", "components", ARCH, "-o", "wide")).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION PROVIDER   REPOSITORY
test.de/x v1      mandelsoft /tmp/ca
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
			Expect(env.CatchOutput(buf).Execute("get", "components", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION PROVIDER
test.de/x v1      mandelsoft
test.de/x v1.1    mandelsoft
test.de/x v2      mandelsoft
`))
		})

		It("lists latest version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "--latest", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION PROVIDER
test.de/x v2      mandelsoft
`))
		})

		It("lists constrainted version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "--constraints", ">1.0", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION PROVIDER
test.de/x v1.1    mandelsoft
test.de/x v2      mandelsoft
`))
		})

		It("lists constrainted version", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "--constraints", "1.x.x", "--latest", "--repo", ARCH, COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION PROVIDER
test.de/x v1.1    mandelsoft
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
			})
			env.OCMCommonTransport(ARCH2, accessio.FormatDirectory, func() {
				env.Component(COMP, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
					})
				})
			})
		})
		It("lists closure ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "--lookup", ARCH2, "-r", "--repo", ARCH, COMP2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
REFERENCEPATH COMPONENT VERSION PROVIDER   IDENTITY
              test.de/y v1      mandelsoft 
test.de/y:v1  test.de/x v1      mandelsoft "name"="xx"
`))
		})
		It("lists flat ctf file", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "-o", "tree", "--repo", ARCH, COMP2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING COMPONENT VERSION PROVIDER
└─      test.de/y v1      mandelsoft
`))
		})
		It("lists flat ctf file with closure", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "-o", "tree", "--lookup", ARCH2, "-r", "--repo", ARCH, COMP2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
NESTING COMPONENT VERSION PROVIDER   IDENTITY
└─ ⊗    test.de/y v1      mandelsoft 
   └─   test.de/x v1      mandelsoft "name"="xx"
`))
		})

		It("lists converted yaml", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("get", "components", "-S", compdescv3.VersionName, "-o", "yaml", "--repo", ARCH, COMP2)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				fmt.Sprintf(`
---
apiVersion: %s
kind: ComponentVersion
metadata:
  name: test.de/y
  provider:
    name: mandelsoft
  version: v1
repositoryContexts: []
spec:
  references:
  - componentName: test.de/x
    name: xx
    version: v1
`, compdescv3.SchemaVersion)))
		})
	})
})
