package cobrautils_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/cobrautils"
)

var _ = Describe("processing buffer", func() {
	Context("cleanup", func() {
		It("handles list with text", func() {
			s := cobrautils.CleanMarkdown(`
this is a description list:
- item1

  some text

- item2

  some text
`)
			Expect(s).To(Equal(`
this is a description list:
- item1
  some text

- item2
  some text
`))
		})

		It("handles list with nested list", func() {
			s := cobrautils.CleanMarkdown(`
this is a description list:
- item1

  - sub list

- item2

  - sub list
`)
			Expect(s).To(StringEqualWithContext(`
this is a description list:
- item1
  - sub list

- item2
  - sub list`))
		})
	})
})
