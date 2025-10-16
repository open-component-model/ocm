package env

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"ocm.software/ocm/api/ocm"
)

var _ = Describe("Environment", func() {
	It("loads environment", func() {
		h := NewEnvironment(TestData())
		defer h.Cleanup()
		data, err := vfs.ReadFile(h.FileSystem(), "/testdata/testfile.txt")
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("this is some test data"))
	})

	It("reuses context", func() {
		ctx := ocm.New()
		h := NewEnvironment(OCMContext(ctx), FileSystem(osfs.OsFs))
		Expect(h.OCMContext()).To(BeIdenticalTo(ctx))
	})
})
