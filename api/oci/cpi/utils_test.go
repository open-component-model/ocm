package cpi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/oci/cpi"
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
