package genericocireg_test

import (
	"encoding/json"
	"reflect"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	ocmreg "ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/utils/runtime"
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
