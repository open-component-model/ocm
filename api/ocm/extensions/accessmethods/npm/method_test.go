package npm_test

import (
	"crypto"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/npm"
	"ocm.software/ocm/api/tech/npm/npmtest"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
)

var _ = Describe("Method", func() {
	var cv ocm.ComponentVersionAccess
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(npmtest.TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("local", func() {
		It("accesses artifact", func() {
			acc := npm.New("file://"+npmtest.NPMPATH, npmtest.PACKAGE, npmtest.VERSION)

			m := Must(acc.AccessMethod(cv))
			defer m.Close()
			Expect(m.MimeType()).To(Equal(mime.MIME_TGZ))

			r := Must(m.Reader())
			defer r.Close()
			dr := iotools.NewDigestReaderWithHash(crypto.SHA1, r)
			for {
				var buf [8096]byte
				_, err := dr.Read(buf[:])
				if err != nil {
					break
				}
			}
			Expect(dr.Size()).To(Equal(int64(npmtest.ARTIFACT_SIZE)))
			Expect(dr.Digest().String()).To(Equal(npmtest.ARTIFACT_DIGEST))
		})

		It("detects digests mismatch", func() {
			acc := npm.New("file://"+npmtest.FAILPATH, npmtest.PACKAGE, npmtest.VERSION)

			m := Must(acc.AccessMethod(cv))
			defer m.Close()
			_, err := m.Reader()
			Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found 34a77645201d1a8fc5213ace787c220eabbd0967")))
		})
	})

	Context("npmjs", func() {
		It("accesses", func() {
			acc := npm.New("https://registry.npmjs.org/", npmtest.PACKAGE, npmtest.VERSION)
			m := Must(acc.AccessMethod(cv))
			defer m.Close()
			data := Must(m.Get())
			Expect(len(data)).To(Equal(npmtest.ARTIFACT_SIZE))
		})
	})
})
