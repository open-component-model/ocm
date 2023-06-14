// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cobrautils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/cobrautils"
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
