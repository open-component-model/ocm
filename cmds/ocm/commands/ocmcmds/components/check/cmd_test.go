package check_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/accessio"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH    = "/tmp/ca"
	VERSION = "v1"
	COMP    = "test.de/x"
	COMP2   = "test.de/y"
	COMP3   = "test.de/z"
	COMP4   = "test.de/a"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get checks references", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION)
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION STATUS ERROR
test.de/x v1      OK
`))
	})

	Context("finds missing", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERSION, func() {
					env.Reference("ref", COMP3, VERSION)
				})
				env.ComponentVersion(COMP2, VERSION, func() {
					env.Reference("ref", COMP3, VERSION)
				})
			})
		})

		It("outputs table", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP)).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION STATUS     ERROR
test.de/x v1      Incomplete
`))
		})

		It("outputs wide table", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "-o", "wide")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION STATUS     ERROR MISSING                    NON-LOCAL
test.de/x v1      Incomplete       test.de/z:v1[test.de/x:v1]
`))
		})

		It("outputs json", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "-o", "json")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
{
  "items": [
    {
      "componentVersion": "test.de/x:v1",
      "missing": {
        "test.de/z:v1": [
          "test.de/x:v1",
          "test.de/z:v1"
        ]
      },
      "status": "Incomplete"
    }
  ]
}
`))
		})

		It("provides error table", func() {
			buf := bytes.NewBuffer(nil)
			ExpectError(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "--fail-on-error")).
				To(MatchError("incomplete component version test.de/x:v1"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION STATUS     ERROR
test.de/x v1      Incomplete
`))
		})
	})

	It("handles diamond", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref1", COMP2, VERSION)
				env.Reference("ref2", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP4, VERSION, func() {
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION STATUS ERROR
test.de/x v1      OK     
`))
	})
	It("handles all", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref1", COMP2, VERSION)
				env.Reference("ref2", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP4, VERSION, func() {
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("check", "components", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION STATUS ERROR
test.de/a v1      OK     
test.de/x v1      OK     
test.de/y v1      OK     
test.de/z v1      OK     
`))
	})

	It("finds cycle", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP2, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION STATUS ERROR
test.de/x v1      Error  component version recursion: use of test.de/z:v1 for test.de/x:v1->test.de/z:v1->test.de/y:v1
`))
	})

	It("finds all cycles", func() {
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMP, VERSION, func() {
				env.Reference("ref", COMP2, VERSION)
				env.Reference("ref", COMP3, VERSION)
			})
			env.ComponentVersion(COMP2, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP3, VERSION, func() {
				env.Reference("ref", COMP4, VERSION)
			})
			env.ComponentVersion(COMP4, VERSION, func() {
				env.Reference("ref", COMP, VERSION)
			})
		})

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("check", "components", ARCH)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
COMPONENT VERSION STATUS ERROR
test.de/a v1      Error  component version recursion: use of test.de/a:v1 for test.de/a:v1->test.de/x:v1->test.de/z:v1
test.de/x v1      Error  component version recursion: use of test.de/x:v1 for test.de/x:v1->test.de/z:v1->test.de/a:v1
test.de/y v1      Error  component version recursion: use of test.de/a:v1 for test.de/y:v1->test.de/a:v1->test.de/x:v1->test.de/z:v1
test.de/z v1      Error  component version recursion: use of test.de/z:v1 for test.de/z:v1->test.de/a:v1->test.de/x:v1
`))
	})

	Context("finds non-local resources", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP, VERSION, func() {
					env.Resource("rsc1", VERSION, resourcetypes.BLUEPRINT, v1.LocalRelation, func() {
						env.ModificationOptions(ocm.SkipDigest())
						env.Access(ociartifact.New("bla"))
					})
				})
			})
		})

		It("outputs table", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "--local-resources")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION STATUS    ERROR
test.de/x v1      Resources
`))
		})

		It("outputs wide table", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "--local-resources", "-o=wide")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
COMPONENT VERSION STATUS    ERROR MISSING NON-LOCAL
test.de/x v1      Resources               RSC("name"="rsc1")
`))
		})
	})
})
