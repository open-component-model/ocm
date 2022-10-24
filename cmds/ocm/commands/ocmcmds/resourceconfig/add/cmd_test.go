// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package add_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/testutils"
)

const SPECFILE = "/tmp/resources.yaml"
const VERSION = "v1"

func CheckSpec(env *TestEnv, spec string) {
	data, err := env.ReadFile(SPECFILE)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, string(data)).To(testutils.StringEqualTrimmedWithContext(spec))

}

var _ = Describe("Add resources", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("resource by options", func() {
		It("adds simple text blob", func() {
			meta := `
name: testdata
type: PlainText
`
			input := `
type: file
path: ../testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "resourceconfig", SPECFILE, "--resource", meta, "--input", input)).To(Succeed())
			CheckSpec(env, `
---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata
type: PlainText
`)
		})

		It("adds a second simple text blob", func() {
			meta1 := `
name: testdata1
type: PlainText
`
			meta2 := `
name: testdata2
type: PlainText
`
			input := `
type: file
path: ../testdata/testcontent
mediaType: text/plain
`
			Expect(env.Execute("add", "resourceconfig", SPECFILE, "--resource", meta1, "--input", input)).To(Succeed())
			Expect(env.Execute("add", "resourceconfig", SPECFILE, "--resource", meta2, "--input", input)).To(Succeed())
			CheckSpec(env, `
---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata1
type: PlainText

---
input:
  mediaType: text/plain
  path: ../testdata/testcontent
  type: file
name: testdata2
type: PlainText
`)
		})
	})
})
