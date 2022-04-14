// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package ocm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/ocm"
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
	Expect(err).To(Succeed())
	Expect(spec).To(Equal(ocm.RefSpec{
		UniformRepositorySpec: ocm.UniformRepositorySpec{
			Type:    ut,
			Host:    h,
			SubPath: us,
			Info:    i,
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
	})
})
