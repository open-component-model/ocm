package ocm_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci/extensions/repositories/empty"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	ocmreg "ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/utils/runtime"
)

var DefaultContext = ocm.New()

var _ = Describe("access method", func() {
	It("instantiate repo mapped to empty oci repo", func() {
		backendSpec := genericocireg.NewRepositorySpec(
			empty.NewRepositorySpec(),
			ocmreg.NewComponentRepositoryMeta("", ocmreg.OCIRegistryDigestMapping))
		data, err := json.Marshal(backendSpec)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"componentNameMapping\":\"sha256-digest\",\"type\":\"Empty\"}"))

		repo, err := DefaultContext.RepositoryForConfig(data, runtime.DefaultYAMLEncoding)
		Expect(err).To(Succeed())
		Expect(repo).NotTo(BeNil())
	})
})
