package grammar

import (
	"fmt"
	"regexp"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Masterminds/semver/v3"
	gr "github.com/mandelsoft/goutils/regexutils"
)

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

func Scheme(sc string) string {
	if sc == "" {
		return sc
	}
	return sc + "://"
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

		It("matches complex semver", func() {
			Expect(VersionRegexp.MatchString("0.2.3+2024.T06b")).To(BeTrue())
		})

		It("matches complex semver with v prefix", func() {
			Expect(VersionRegexp.MatchString("v0.2.3+2024.T06b")).To(BeTrue())
		})

		for _, pre := range []string{"", "alpha1", "alpha.1.2", "alpha-1"} {
			for _, build := range []string{"", "2024", "2024.1.T2b", "2024.1-T2b"} {
				suf := ""
				if pre != "" {
					suf += "-" + pre
				}
				if build != "" {
					suf += "+" + build
				}
				It(fmt.Sprintf("handles semver %s", suf), func() {
					v := "v0.2.3" + suf
					Must(semver.NewVersion(v))
					Expect(VersionRegexp.MatchString(v)).To(BeTrue())
				})
			}
		}
	})

	Context("complete refs", func() {
		t := "OCIRepository"
		sc := "http"
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		It("succeeds", func() {
			for _, ut := range []string{"", t} {
				for _, usc := range []string{"", sc} {
					for _, us := range []string{"", s} {
						for _, uv := range []string{"", v} {
							ref := Type(ut) + Scheme(usc) + h + Sub(us) + "//" + c + Vers(uv)
							CheckRef(ref, ut, usc, h, us, c, uv)
						}
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
			Check("directory::ghcr.io/sub/path", AnchoredRepositoryRegexp, "directory", "", "ghcr.io", "sub/path")
			Check("ghcr.io/sub/path", AnchoredRepositoryRegexp, "", "", "ghcr.io", "sub/path")
			Check("ghcr.io", AnchoredRepositoryRegexp, "", "", "ghcr.io", "")
			Check("ghcr.io/sub/path", AnchoredRepositoryRegexp, "", "", "ghcr.io", "sub/path")
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
