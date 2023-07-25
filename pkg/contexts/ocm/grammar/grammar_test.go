// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	gr "github.com/open-component-model/ocm/v2/pkg/regex"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCI Test Suite")
}

func CheckRef(ref string, parts ...string) {
	CheckWithOffset(1, ref, AnchoredReferenceRegexp, parts...)
}

func CheckVers(ref string, parts ...string) {
	CheckWithOffset(1, ref, AnchoredComponentVersionRegexp, parts...)
}

func Check(ref string, exp *regexp.Regexp, parts ...string) {
	CheckWithOffset(1, ref, exp, parts...)
}

func CheckWithOffset(o int, ref string, exp *regexp.Regexp, parts ...string) {
	spec := exp.FindSubmatch([]byte(ref))
	if len(parts) == 0 {
		ExpectWithOffset(o+1, spec).To(BeNil())
	} else {
		result := make([]string, len(spec))
		for i, v := range spec {
			result[i] = string(v)
		}
		ExpectWithOffset(o+1, result).To(Equal(append([]string{ref}, parts...)))
	}
}

func Type(t string) string {
	if t == "" {
		return t
	}
	return t + "::"
}
func Sub(t string) string {
	if t == "" {
		return t
	}
	return "/" + t
}
func Vers(t string) string {
	if t == "" {
		return t
	}
	return ":" + t
}

var _ = Describe("ref matching", func() {
	Context("basic", func() {
		It("matches", func() {
			Expect(gr.Match("[a^-]").MatchString("a^--")).To(BeTrue())
			Expect(gr.Match(`[a\-^]`).MatchString("a^-")).To(BeTrue())
			Expect(gr.Match(`[\^a\-]`).MatchString("a^-")).To(BeTrue())
			Expect(gr.Match("[^a^-]").MatchString("b")).To(BeTrue())
			Expect(gr.Match("[^a^-]").MatchString("^")).To(BeFalse())
		})
	})

	Context("versions", func() {
		It("matches versions", func() {
			Expect(VersionRegexp.MatchString("v1.1.1")).To(BeTrue())
			Expect(VersionRegexp.MatchString("v1")).To(BeTrue())
			Expect(VersionRegexp.MatchString("1.1.1")).To(BeTrue())
		})

		It("matches pre versions", func() {
			Expect(VersionRegexp.MatchString("v1.1.1-rc.1")).To(BeTrue())
			Expect(VersionRegexp.MatchString("v1-rc.1")).To(BeTrue())
			Expect(VersionRegexp.MatchString("1.1.1-rc.1")).To(BeTrue())
		})
	})

	Context("complete refs", func() {
		t := "OCIRepository"
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		It("succeeds", func() {
			for _, ut := range []string{"", t} {
				for _, us := range []string{"", s} {
					for _, uv := range []string{"", v} {
						ref := Type(ut) + h + Sub(us) + "//" + c + Vers(uv)
						CheckRef(ref, ut, h, us, c, uv)
					}
				}
			}
		})
		It("fails", func() {
			CheckRef("ghcr/sub//comp")
			CheckRef("ghcr/sub//comp.io/comp")
			CheckRef("ghcr.io/sub/comp.io/comp")
			CheckRef("T:ghcr.io/sub//comp.io/comp")

			CheckRef("directory::./some/../path.dir")
			CheckRef("directory::./some/../path.dir//github.com/mandelsoft/kubelink")
		})
	})

	Context("generic", func() {
		It("succeeds", func() {
			Check("directory::./some/../path.dir", AnchoredGenericReferenceRegexp, "directory", "./some/../path.dir", "", "")
			Check("directory::./some/../path.dir//github.com/mandelsoft/kubelink", AnchoredGenericReferenceRegexp, "directory", "./some/../path.dir", "github.com/mandelsoft/kubelink", "")
			Check("directory::./some/../path.dir//github.com/mandelsoft/kubelink:v1", AnchoredGenericReferenceRegexp, "directory", "./some/../path.dir", "github.com/mandelsoft/kubelink", "v1")
		})
		It("fails", func() {
		})
	})

	Context("repo", func() {
		It("succeeds", func() {
			Check("directory::ghcr.io/sub/path", AnchoredRepositoryRegexp, "directory", "ghcr.io", "sub/path")
			Check("ghcr.io/sub/path", AnchoredRepositoryRegexp, "", "ghcr.io", "sub/path")
			Check("ghcr.io", AnchoredRepositoryRegexp, "", "ghcr.io", "")
			Check("ghcr.io/sub/path", AnchoredRepositoryRegexp, "", "ghcr.io", "sub/path")
		})
		It("fails", func() {
			Check("/ghcr.io/sub/path", AnchoredRepositoryRegexp)
		})
	})

	Context("generic repo", func() {
		It("succeeds", func() {
			Check("directory::./some/../path.dir", AnchoredGenericRepositoryRegexp, "directory", "./some/../path.dir")
			Check("./some/../path.dir", AnchoredGenericRepositoryRegexp, "", "./some/../path.dir")
		})
		It("fails", func() {
		})
	})

	Context("components", func() {
		It("succeeds", func() {
			CheckVers("ghcr.io/test", "ghcr.io/test", "")
			CheckVers("ghcr.io/test:v1", "ghcr.io/test", "v1")
			CheckVers("ghcr.io/test:v1-rc1", "ghcr.io/test", "v1-rc1")
			CheckVers("ghcr.io/test:v1+rc1", "ghcr.io/test", "v1+rc1")
			CheckVers("ghcr.io/test:v1-rc1+build", "ghcr.io/test", "v1-rc1+build")
			CheckVers("ghcr.io/test:v1.1.1-rc1+build", "ghcr.io/test", "v1.1.1-rc1+build")
			CheckVers("ghcr.io/test:v1.1-rc1+build", "ghcr.io/test", "v1.1-rc1+build")
		})
		It("fails", func() {
			CheckVers("ghcr/test")
			CheckVers("ghcr.io:v1")
			CheckVers(":v1")
		})

	})

})
