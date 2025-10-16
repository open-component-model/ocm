package genericocireg_test

import (
	"github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm"
)

var _ = Describe("decode fallback", func() {
	It("creates a dummy component", func() {
		specdata := `
type: other/v1
subPath: test
other: value
`
		spec := testutils.Must(DefaultContext.RepositoryTypes().Decode([]byte(specdata), nil))
		Expect(ocm.IsUnknownRepositorySpec(spec.(ocm.RepositorySpec))).To(BeTrue())
	})
})
