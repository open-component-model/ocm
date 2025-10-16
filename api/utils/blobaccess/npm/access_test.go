package npm_test

import (
	"crypto"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/tech/npm/npmtest"
	"ocm.software/ocm/api/utils/blobaccess/npm"
	"ocm.software/ocm/api/utils/iotools"
	"ocm.software/ocm/api/utils/mime"
)

var _ = Describe("Method", func() {
	It("PackageUrl()", func() {
		packageUrl := "https://registry.npmjs.org/yargs"
		acc := Must(npm.NewPackageSpec("https://registry.npmjs.org", "yargs", "17.7.1"))
		Expect(acc.PackageUrl()).To(Equal(packageUrl))
		acc = Must(npm.NewPackageSpec("https://registry.npmjs.org/", "yargs", "17.7.1"))
		Expect(acc.PackageUrl()).To(Equal(packageUrl))
	})

	It("PackageVersionUrl()", func() {
		packageVersionUrl := "https://registry.npmjs.org/yargs/17.7.1"
		acc := Must(npm.NewPackageSpec("https://registry.npmjs.org", "yargs", "17.7.1"))
		Expect(acc.PackageVersionUrl()).To(Equal(packageVersionUrl))
		acc = Must(npm.NewPackageSpec("https://registry.npmjs.org/", "yargs", "17.7.1"))
		Expect(acc.PackageVersionUrl()).To(Equal(packageVersionUrl))
	})

	Context("access", func() {
		var env *Builder

		BeforeEach(func() {
			env = NewBuilder(npmtest.TestData())
		})

		AfterEach(func() {
			npm.NewPackageSpec("", "", "", npm.WithCredentialContext(nil))
			env.Cleanup()
		})

		It("accesses artifact", func() {
			acc := Must(npm.BlobAccess("file://"+npmtest.NPMPATH, npmtest.PACKAGE, npmtest.VERSION, npm.WithPathFileSystem(env.FileSystem())))
			defer acc.Close()
			Expect(acc.MimeType()).To(Equal(mime.MIME_TGZ))

			r := Must(acc.Reader())
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
			acc := Must(npm.BlobAccess("file://"+npmtest.FAILPATH, npmtest.PACKAGE, npmtest.VERSION, npm.WithPathFileSystem(env.FileSystem())))
			defer acc.Close()
			_, err := acc.Reader()
			Expect(err).To(MatchError(ContainSubstring("SHA-1 digest mismatch: expected 44a77645201d1a8fc5213ace787c220eabbd0967, found 34a77645201d1a8fc5213ace787c220eabbd0967")))
		})
	})

	Context("npmjs", func() {
		It("accesses", func() {
			acc := Must(npm.BlobAccess("https://registry.npmjs.org/", npmtest.PACKAGE, npmtest.VERSION))

			data := Must(acc.Get())
			Expect(len(data)).To(Equal(npmtest.ARTIFACT_SIZE))
		})
	})
})
