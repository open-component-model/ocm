// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cpi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/cpi"
)

var _ = Describe("OCI CPI utils", func() {
	list := []string{
		"a/b/c/d",
		"a/b/c",
		"a/b",
		"a/c/d",
		"a/c",
		"b/c",
	}

	It("calculates exclusive number", func() {
		Expect(cpi.FilterByNamespacePrefix("a/b/", list)).To(Equal([]string{
			"a/b/c/d",
			"a/b/c",
		}))
	})
	It("calculates inclusive number", func() {
		Expect(cpi.FilterByNamespacePrefix("a/b", list)).To(Equal([]string{
			"a/b/c/d",
			"a/b/c",
			"a/b",
		}))
	})

	It("calculates closure", func() {
		Expect(cpi.FilterChildren(true, "a/b", list)).To(Equal([]string{
			"a/b/c/d",
			"a/b/c",
			"a/b",
		}))
	})

	It("calculates children", func() {
		Expect(cpi.FilterChildren(false, "a/b/", list)).To(Equal([]string{
			"a/b/c",
		}))
	})

	It("calculates inclusive children", func() {
		Expect(cpi.FilterChildren(false, "a/b", list)).To(Equal([]string{
			"a/b/c",
			"a/b",
		}))
	})

	It("calculates children closure", func() {
		Expect(cpi.FilterChildren(true, "a/b", cpi.FilterByNamespacePrefix("a/b", list))).To(Equal([]string{
			"a/b/c/d",
			"a/b/c",
			"a/b",
		}))
	})

})
