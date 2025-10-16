package ocireg_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
)

var _ = Describe("ref parsing", func() {
	Context("deserialization", func() {
		It("handles regular spec", func() {
			spec := `
type: ` + ocireg.Type + `
baseUrl: ghcr.io
subPath: open-component-model/ocm
`
			s := Must(ocm.DefaultContext().RepositorySpecForConfig([]byte(spec), nil))

			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")))
		})

		It("handles combined url", func() {
			spec := `
type: ` + ocireg.Type + `
baseUrl: ghcr.io/open-component-model/ocm
`
			s := Must(ocm.DefaultContext().RepositorySpecForConfig([]byte(spec), nil))

			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")))
		})

		It("handles scheme", func() {
			spec := `
type: ` + ocireg.Type + `
baseUrl: https://ghcr.io/open-component-model/ocm
`
			s := Must(ocm.DefaultContext().RepositorySpecForConfig([]byte(spec), nil))

			Expect(s).To(Equal(ocireg.NewRepositorySpec("https://ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("https://ghcr.io/open-component-model/ocm")))
		})
	})

	Context("constructor", func() {
		It("handles path", func() {
			s := ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")))
		})

		It("handles scheme", func() {
			s := ocireg.NewRepositorySpec("https://ghcr.io/open-component-model/ocm")
			Expect(s).To(Equal(ocireg.NewRepositorySpec("https://ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("https://ghcr.io/open-component-model/ocm")))
		})

		It("handles meta", func() {
			s := ocireg.NewRepositorySpec("ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io", ocireg.NewComponentRepositoryMeta("open-component-model/ocm"))))
			Expect(s).To(Equal(ocireg.NewRepositorySpec("ghcr.io/open-component-model/ocm")))
		})
	})
})
