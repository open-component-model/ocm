// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package common_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common"
)

var _ = Describe("Blob Inputs", func() {

	It("missing input", func() {
		in := `
access:
  type: localBlob
`
		_, err := common.DecodeInput([]byte(in), nil)
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
		_, err := common.DecodeInput([]byte(in), nil)
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
		_, err := common.DecodeInput([]byte(in), nil)
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
		_, err := common.DecodeInput([]byte(in), nil)
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
		_, err := common.DecodeInput([]byte(in), nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("input.excludeFiles: Forbidden: unknown field"))
	})
})
