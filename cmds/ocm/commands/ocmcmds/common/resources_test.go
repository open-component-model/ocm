package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
)

var _ = Describe("Blob Inputs", func() {
	It("missing input", func() {
		in := `
access:
  type: localBlob
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("simple decode", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})
	It("complains about additional input field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  bla: blub
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err.Error()).To(Equal("input.bla: Forbidden: unknown field"))
	})

	It("does not complains about additional dir field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: dir
  excludeFiles:
     - xyz
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("complains about additional dir field for file", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  excludeFiles:
  - xyz
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("input.excludeFiles: Forbidden: unknown field"))
	})
})
