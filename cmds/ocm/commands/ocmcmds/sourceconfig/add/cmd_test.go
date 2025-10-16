package add_test

import (
	"github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	SPECFILE = "/tmp/sources.yaml"
	VERSION  = "v1"
)

func CheckSpec(env *TestEnv, spec string) {
	data, err := env.ReadFile(SPECFILE)
	ExpectWithOffset(1, err).To(Succeed())
	ExpectWithOffset(1, string(data)).To(testutils.StringEqualTrimmedWithContext(spec))
}

var _ = Describe("Add sources", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("source by options", func() {
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
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta, "--input", input)).To(Succeed())
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

		It("defaults artifact type", func() {
			access := `
type: gitHub
repoUrl: github.com/open-component-model/ocm
commit: xxx
`
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--name", "sources", "--access", access)).To(Succeed())
			CheckSpec(env, `
---
access:
  commit: xxx
  repoUrl: github.com/open-component-model/ocm
  type: gitHub
name: sources
type: directoryTree
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
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta1, "--input", input)).To(Succeed())
			Expect(env.Execute("add", "sourceconfig", SPECFILE, "--source", meta2, "--input", input)).To(Succeed())
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
