package ctf_test

import (
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext"
	"github.com/open-component-model/ocm/pkg/contexts/datacontext/attrs/vfsattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	. "github.com/open-component-model/ocm/pkg/testutils"
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
