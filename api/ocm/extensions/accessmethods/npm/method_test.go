package npm_test

import (
	"crypto"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/helper/env"

	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/npm"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
)

const (
	NPMPATH  = "/testdata/registry"
	FAILPATH = "/testdata/failregistry"
)

var _ = Describe("Method", func() {
	var env *Builder
	var cv ocm.ComponentVersionAccess

	BeforeEach(func() {
		env = NewBuilder(TestData())
		cv = &cpi.DummyComponentVersionAccess{env.OCMContext()}
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artifact", func() {
		acc := npm.New("file://"+NPMPATH, "yargs", "17.7.1")
		// acc := npm.New("https://registry.npmjs.org", "yargs", "17.7.1")

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
		Expect(dr.Size()).To(Equal(int64(65690)))
		Expect(dr.Digest().String()).To(Equal("SHA-1:34a77645201d1a8fc5213ace787c220eabbd0967"))
	})

	It("detects digests mismatch", func() {
		acc := npm.New("file://"+FAILPATH, "yargs", "17.7.1")

		m := Must(acc.AccessMethod(cv))
		defer m.Close()
		_, err := m.Reader()
		Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found 34a77645201d1a8fc5213ace787c220eabbd0967")))
	})

	It("PackageUrl()", func() {
		packageUrl := "https://registry.npmjs.org/yargs"
		acc := npm.New("https://registry.npmjs.org", "yargs", "17.7.1")
		Expect(acc.PackageUrl()).To(Equal(packageUrl))
		acc = npm.New("https://registry.npmjs.org/", "yargs", "17.7.1")
		Expect(acc.PackageUrl()).To(Equal(packageUrl))
	})

	It("PackageVersionUrl()", func() {
		packageVersionUrl := "https://registry.npmjs.org/yargs/17.7.1"
		acc := npm.New("https://registry.npmjs.org", "yargs", "17.7.1")
		Expect(acc.PackageVersionUrl()).To(Equal(packageVersionUrl))
		acc = npm.New("https://registry.npmjs.org/", "yargs", "17.7.1")
		Expect(acc.PackageVersionUrl()).To(Equal(packageVersionUrl))
	})
})
