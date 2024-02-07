// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package check_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/common/accessio"
)

const ARCH = "/tmp/ca"
const VERSION = "v1"
const COMP = "test.de/x"
const COMP2 = "test.de/y"
const COMP3 = "test.de/z"
const COMP4 = "test.de/a"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("get checks refereces", func() {
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
		It("outputs json", func() {
			buf := bytes.NewBuffer(nil)
			Expect(env.CatchOutput(buf).Execute("check", "components", ARCH+"//"+COMP, "-o", "json")).To(Succeed())
			Expect(buf.String()).To(StringEqualTrimmedWithContext(
				`
{
  "items": [
    {
      "status": "Incomplete",
      "componentVersion": "test.de/x:v1",
      "missing": {
        "test.de/z:v1": [
          "test.de/x:v1",
          "test.de/z:v1"
        ]
      }
    }
  ]
}
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
})
