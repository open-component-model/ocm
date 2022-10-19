// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
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

func CheckRef(ref, ut, h, us, c, uv, i string) {
	var v *string
	if uv != "" {
		v = &uv
	}
	spec, err := ocm.ParseRef(ref)
	Expect(err).WithOffset(1).To(Succeed())
	Expect(spec).WithOffset(1).To(Equal(ocm.RefSpec{
		UniformRepositorySpec: ocm.UniformRepositorySpec{
			Type:            ut,
			Host:            h,
			SubPath:         us,
			Info:            i,
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
		t := "OCIRepository"
		s := "mandelsoft/cnudie"
		v := "v1"

		h := "ghcr.io"
		c := "github.com/mandelsoft/ocm"

		It("without info", func() {
			for _, ut := range []string{"", t} {
				for _, us := range []string{"", s} {
					for _, uv := range []string{"", v} {
						ref := Type(ut) + h + Sub(us) + "//" + c + Vers(uv)
						CheckRef(ref, ut, h, us, c, uv, "")
					}
				}
			}
		})

		It("info", func() {
			for _, ut := range []string{"", t} {
				CheckRef(Type(ut)+"{}", ut, "", "", "", "", "{}")
			}
		})

		It("info+comp", func() {
			for _, ut := range []string{"", t} {
				for _, uv := range []string{"", v} {
					CheckRef(Type(ut)+"{}"+"//"+c+Vers(uv), ut, "", "", c, uv, "{}")
				}
			}
		})

		It("dir ref", func() {
			CheckRef("+ctf+directory::./file//bla.blob/comp", "ctf+directory", "", "", "bla.blob/comp", "", "./file")
			CheckRef("ctf+directory::./file//bla.blob/comp", "ctf+directory", "", "", "bla.blob/comp", "", "./file")
			CheckRef("directory::./file//bla.blob/comp", "directory", "", "", "bla.blob/comp", "", "./file")
			CheckRef("directory::file//bla.blob/comp", "directory", "", "", "bla.blob/comp", "", "file")
			CheckRef("directory::./file.io//bla.blob/comp", "directory", "", "", "bla.blob/comp", "", "./file.io")
			CheckRef("any::file.io//bla.blob/comp", "any", "file.io", "", "bla.blob/comp", "", "")
		})
	})
})
