// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/pkg/utils"
)

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

func CheckRef(ref, ut, scheme, h, us, c, uv, i string, th ...string) {
	var v *string
	if uv != "" {
		v = &uv
	}
	if len(th) == 0 && ut != "" {
		th = []string{ut}
	}
	spec, err := ocm.ParseRef(ref)
	Expect(err).WithOffset(1).To(Succeed())
	Expect(spec).WithOffset(1).To(Equal(ocm.RefSpec{
		UniformRepositorySpec: ocm.UniformRepositorySpec{
			Type:            ut,
			Scheme:          scheme,
			Host:            h,
			SubPath:         us,
			Info:            i,
			TypeHint:        utils.Optional(th...),
			CreateIfMissing: ref[0] == '+',
		},
		CompSpec: ocm.CompSpec{
			Component: c,
			Version:   v,
		},
	}))
}

var _ = Describe("ref parsing", func() {
	Context("complete refs", func() {
		t := ocireg.Type
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		Context("without info", func() {
			for _, ut := range []string{t, ""} {
				for _, uh := range []string{h, h + ":3030", "localhost", "localhost:3030"} {
					for _, us := range []string{"", s} {
						for _, uv := range []string{"", v, v + ".1.1", v + "-rc.1", v + "+65", v + ".1.2-rc.1", v + ".1.2+65"} {
							ref := Type(ut) + uh + Sub(us) + "//" + c + Vers(uv)
							ut, uh, us, uv := ut, uh, us, uv

							It("parses ref "+ref, func() {
								if ut == "" && strings.HasPrefix(uh, "localhost") {
									CheckRef(ref, ut, "", "", "", c, uv, uh+Sub(us))
								} else {
									CheckRef(ref, ut, "", uh, us, c, uv, "")
								}
							})
						}
					}
				}
			}
		})

		It("scheme", func() {
			CheckRef("OCIRegistry::http://ghcr.io:80/repository//acme.org/component:1.0.0", "OCIRegistry", "http", "ghcr.io:80", "repository", "acme.org/component", "1.0.0", "")
		})

		It("info", func() {
			for _, ut := range []string{"", t} {
				CheckRef(Type(ut)+"{}", ut, "", "", "", "", "", "{}")
			}
		})

		It("info+comp", func() {
			for _, ut := range []string{"", t} {
				for _, uv := range []string{"", v} {
					CheckRef(Type(ut)+"{}"+"//"+c+Vers(uv), ut, "", "", "", c, uv, "{}")
				}
			}
		})

		It("dir ref", func() {
			CheckRef("+ctf+directory::./file//bla.blob/comp", "ctf", "", "", "", "bla.blob/comp", "", "./file", "ctf+directory")
			CheckRef("ctf+directory::./file//bla.blob/comp", "ctf", "", "", "", "bla.blob/comp", "", "./file", "ctf+directory")
			CheckRef("directory::./file//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "./file")
			CheckRef("directory::file//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "file")
			CheckRef("directory::./file.io//bla.blob/comp", "directory", "", "", "", "bla.blob/comp", "", "./file.io")
			CheckRef("any::file.io//bla.blob/comp", "any", "", "file.io", "", "bla.blob/comp", "", "")
		})

		It("dedicated test case", func() {
			ref := Must(ocm.ParseRef("OCIRegistry::73555000100900003892.dev.dockersrv.repositories.sapcloud.cn/dev/v1//github.wdf.sap.corp/kubernetes/landscape-setup-dependencies:0.3797.0"))

			v := "0.3797.0"
			Expect(ref).To(Equal(ocm.RefSpec{
				UniformRepositorySpec: ocm.UniformRepositorySpec{
					Type:            "OCIRegistry",
					Host:            "73555000100900003892.dev.dockersrv.repositories.sapcloud.cn",
					SubPath:         "dev/v1",
					Info:            "",
					CreateIfMissing: false,
					TypeHint:        "OCIRegistry",
				},
				CompSpec: ocm.CompSpec{
					Component: "github.wdf.sap.corp/kubernetes/landscape-setup-dependencies",
					Version:   &v,
				},
			}))
		})
	})

	Context("map to spec", func() {
		It("handles localhost", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("OCIRegistry::http://localhost:80/test//github.vom/mandelsoft/test"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			Expect(spec).To(Equal(ocireg.NewRepositorySpec("http://localhost:80", ocireg.NewComponentRepositoryMeta("test", ""))))
			repo := Must(spec.Repository(ctx, nil))
			specFromRepo := repo.GetSpecification()
			Expect(spec).To(Equal(specFromRepo))
		})
		It("handles localhost without scheme", func() {
			ctx := ocm.New()

			ref := Must(ocm.ParseRef("OCIRegistry::localhost:80/test//github.vom/mandelsoft/test"))
			spec := Must(ctx.MapUniformRepositorySpec(&ref.UniformRepositorySpec))
			Expect(spec).To(Equal(ocireg.NewRepositorySpec("localhost:80", ocireg.NewComponentRepositoryMeta("test", ""))))
			repo := Must(spec.Repository(ctx, nil))
			specFromRepo := repo.GetSpecification()
			Expect(spec).To(Equal(specFromRepo))
		})
	})
})
