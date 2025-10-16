package ctf_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/datacontext"
	"ocm.software/ocm/api/datacontext/attrs/vfsattr"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
)

var _ = Describe("access method", func() {
	var fs vfs.FileSystem
	var ctx ocm.Context

	BeforeEach(func() {
		ctx = ocm.New(datacontext.MODE_EXTENDED)
		fs = memoryfs.New()
		vfsattr.Set(ctx.AttributesContext(), fs)
	})

	Context("create", func() {
		It("create ctf", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "+ctf::test/repository"))
			Expect(ctf.NewRepositorySpec(ctf.ACC_CREATE, "test/repository", accessio.FormatDirectory, accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})

		It("create directory", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "+directory::test/repository"))
			Expect(ctf.NewRepositorySpec(ctf.ACC_CREATE, "test/repository", accessio.FormatDirectory, accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})

		It("create tgz", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "+tgz::test/repository"))
			Expect(ctf.NewRepositorySpec(ctf.ACC_CREATE, "test/repository", accessio.FormatTGZ, accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})

		It("create ca", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "+ca::test/repository"))
			Expect(comparch.NewRepositorySpec(ctf.ACC_CREATE, "test/repository", accessio.FormatDirectory, accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})
	})

	Context("read", func() {
		It("read ctf", func() {
			ExpectError(ocm.ParseRepoToSpec(ctx, "test/repository")).To(MatchError(`repository specification "test/repository" is invalid: repository "test/repository" is unknown`))
		})

		It("read ctf", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "ctf::test/repository"))
			Expect(ctf.NewRepositorySpec(ctf.ACC_WRITABLE, "test/repository", accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})

		It("read ctf", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "tgz::test/repository"))
			Expect(ctf.NewRepositorySpec(ctf.ACC_WRITABLE, "test/repository", accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})

		It("read ca", func() {
			MustBeSuccessful(fs.Mkdir("test", 0o700))

			spec := Must(ocm.ParseRepoToSpec(ctx, "ca::test/repository"))
			Expect(comparch.NewRepositorySpec(ctf.ACC_WRITABLE, "test/repository", accessio.PathFileSystem(fs))).To(DeepEqual(spec))
		})
	})
})
