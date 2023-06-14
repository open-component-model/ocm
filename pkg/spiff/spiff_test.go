// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package spiff_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/spiff/spiffing"

	"github.com/open-component-model/ocm/pkg/spiff"
)

var _ = Describe("spiff", func() {
	It("calls spiff", func() {
		res := Must(spiff.CascadeWith(spiff.TemplateData("test", []byte("((  \"alice+\" \"bob\" ))")), spiff.Mode(spiffing.MODE_PRIVATE)))
		Expect(string(res)).To(Equal("alice+bob\n"))
	})
})
