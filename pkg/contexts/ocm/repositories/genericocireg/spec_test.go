// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericocireg_test

import (
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/open-component-model/ocm/v2/pkg/contexts/oci"
	"github.com/open-component-model/ocm/v2/pkg/contexts/oci/repositories/ocireg"
	"github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/genericocireg"
	ocmreg "github.com/open-component-model/ocm/v2/pkg/contexts/ocm/repositories/ocireg"
	"github.com/open-component-model/ocm/v2/pkg/runtime"
)

var DefaultOCIContext = oci.New()

var _ = Describe("access method", func() {
	specData := "{\"baseUrl\":\"X\",\"componentNameMapping\":\"sha256-digest\",\"type\":\"OCIRegistry\"}"

	It("marshal mapped spec", func() {
		gen := genericocireg.NewRepositorySpec(
			ocireg.NewRepositorySpec("X"),
			ocmreg.NewComponentRepositoryMeta("", ocmreg.OCIRegistryDigestMapping))
		data, err := json.Marshal(gen)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal(specData))
	})

	It("decodes generic spec", func() {
		del := genericocireg.New(10)

		spec := Must(del.Decode(DefaultContext, []byte(specData), runtime.DefaultJSONEncoding))
		Expect(reflect.TypeOf(spec).String()).To(Equal("*genericocireg.RepositorySpec"))

		eff, ok := spec.(*genericocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(reflect.TypeOf(eff.RepositorySpec).String()).To(Equal("*ocireg.RepositorySpec"))
		Expect(eff.ComponentNameMapping).To(Equal(ocmreg.OCIRegistryDigestMapping))

		Expect(spec.GetType()).To(Equal(ocireg.Type))
		effoci, ok := eff.RepositorySpec.(*ocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(effoci.BaseURL).To(Equal("X"))
	})

	It("decodes generic spec", func() {
		spec := Must(DefaultContext.RepositorySpecForConfig([]byte(specData), nil))

		Expect(reflect.TypeOf(spec).String()).To(Equal("*genericocireg.RepositorySpec"))

		eff, ok := spec.(*genericocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(reflect.TypeOf(eff.RepositorySpec).String()).To(Equal("*ocireg.RepositorySpec"))
		Expect(eff.ComponentNameMapping).To(Equal(ocmreg.OCIRegistryDigestMapping))

		Expect(spec.GetType()).To(Equal(ocireg.Type))
		effoci, ok := eff.RepositorySpec.(*ocireg.RepositorySpec)
		Expect(ok).To(BeTrue())
		Expect(effoci.BaseURL).To(Equal("X"))
	})

	It("creates spec", func() {
		spec := ocmreg.NewRepositorySpec("http://127.0.0.1:5000/ocm")
		Expect(spec).To(Equal(&ocmreg.RepositorySpec{
			RepositorySpec: ocireg.NewRepositorySpec("http://127.0.0.1:5000"),
			ComponentRepositoryMeta: genericocireg.ComponentRepositoryMeta{
				SubPath:              "ocm",
				ComponentNameMapping: genericocireg.OCIRegistryURLPathMapping,
			},
		}))
	})
})
