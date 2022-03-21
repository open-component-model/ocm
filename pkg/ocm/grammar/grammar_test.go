// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package grammar

import (
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCI Test Suite")
}

func CheckRef(ref string, parts ...string) {
	Check(ref, AnchoredReferenceRegexp, parts...)
}

func CheckVers(ref string, parts ...string) {
	Check(ref, AnchoredComponentVersionRegexp, parts...)
}

func Check(ref string, exp *regexp.Regexp, parts ...string) {
	spec := exp.FindSubmatch([]byte(ref))
	if len(parts) == 0 {
		Expect(spec).To(BeNil())
	} else {
		result := make([]string, len(spec))
		for i, v := range spec {
			result[i] = string(v)
		}
		Expect(result).To(Equal(append([]string{ref}, parts...)))
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
		})
		It("fails", func() {
			CheckVers("ghcr/test")
			CheckVers("ghcr.io:v1")
			CheckVers(":v1")
		})

	})

})
