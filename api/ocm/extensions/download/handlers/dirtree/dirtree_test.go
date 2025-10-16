package dirtree_test

import (
	"encoding/json"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"ocm.software/ocm/api/helper/builder"
	envhelper "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/dirtree"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/tarutils"
)

const (
	COMPONENT = "mandelsoft.org/dirtreeartifact"
	VERSION   = "v1"
	RESOURCE  = "archive"
)

var _ = Describe("artifact management", func() {
	var env *builder.Builder

	cfg := Must(json.Marshal(ociv1.ImageConfig{}))

	BeforeEach(func() {
		env = builder.NewBuilder(envhelper.TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("archive", func() {
		BeforeEach(func() {
			MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/all")), "archive", tarutils.Gzip, env))

			env.OCMCommonTransport("ctf", accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENT, VERSION, func() {
					env.Resource(RESOURCE, VERSION, resourcetypes.DIRECTORY_TREE, metav1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(mime.MIME_TGZ_ALT), "archive")
					})
				})
			})
		})

		It("downloads archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig)

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(h.Download(p, res, "result", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("downloads archive to archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig).SetArchiveMode(true)

			p, buf := common.NewBufferedPrinter()
			accepted, path, err := h.Download(p, res, "target", env)
			Expect(err).To(Succeed())
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("target"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
target: 3584 byte(s) written
`))

			MustBeSuccessful(env.MkdirAll("result", 0o700))
			resultfs := Must(projectionfs.New(env, "result"))
			MustBeSuccessful(tarutils.ExtractArchiveToFs(resultfs, "target", env))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("uses archive downloader", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			p, buf := common.NewBufferedPrinter()
			accepted, path, err := download.For(env).Download(p, res, "result", env)
			Expect(err).To(Succeed())
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})
	})

	Context("single layer", func() {
		BeforeEach(func() {
			MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/all")), "layer0.tgz", tarutils.Gzip, env))

			env.ArtifactSet("image.set", accessio.FormatTGZ, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobData(ociv1.MediaTypeImageConfig, cfg)
					})
					env.Layer(func() {
						env.BlobFromFile(ociv1.MediaTypeImageLayerGzip, "layer0.tgz")
					})
				})
				env.Annotation(artifactset.MAINARTIFACT_ANNOTATION, VERSION)
			})

			env.OCMCommonTransport("ctf", accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENT, VERSION, func() {
					env.Resource(RESOURCE, VERSION, resourcetypes.DIRECTORY_TREE, metav1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(ociv1.MediaTypeImageManifest), "image.set")
					})
				})
			})
		})

		It("downloads archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig)

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(h.Download(p, res, "result", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("downloads archive to archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig).SetArchiveMode(true)

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(h.Download(p, res, "target", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("target"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
target: 3584 byte(s) written
`))

			MustBeSuccessful(env.MkdirAll("result", 0o700))
			resultfs := Must(projectionfs.New(env, "result"))
			MustBeSuccessful(tarutils.ExtractArchiveToFs(resultfs, "target", env))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("uses archive downloader", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(download.For(env).Download(p, res, "result", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})
	})

	Context("multi layer", func() {
		BeforeEach(func() {
			MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/0")), "layer0.tgz", tarutils.Gzip, env))
			MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, "testdata/layers/1")), "layer1.tgz", tarutils.Gzip, env))

			env.ArtifactSet("image.set", accessio.FormatTGZ, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobData(ociv1.MediaTypeImageConfig, cfg)
					})
					env.Layer(func() {
						env.BlobFromFile(ociv1.MediaTypeImageLayerGzip, "layer0.tgz")
					})
					env.Layer(func() {
						env.BlobFromFile(ociv1.MediaTypeImageLayerGzip, "layer1.tgz")
					})
				})
				env.Annotation(artifactset.MAINARTIFACT_ANNOTATION, VERSION)
			})

			env.OCMCommonTransport("ctf", accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENT, VERSION, func() {
					env.Resource(RESOURCE, VERSION, resourcetypes.DIRECTORY_TREE, metav1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(ociv1.MediaTypeImageManifest), "image.set")
					})
				})
			})
		})
		It("downloads archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig)

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(h.Download(p, res, "result", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("downloads archive to archive", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			h := dirtree.New(ociv1.MediaTypeImageConfig).SetArchiveMode(true)

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(h.Download(p, res, "target", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("target"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
target: 3584 byte(s) written
`))

			MustBeSuccessful(env.MkdirAll("result", 0o700))
			resultfs := Must(projectionfs.New(env, "result"))
			MustBeSuccessful(tarutils.ExtractArchiveToFs(resultfs, "target", env))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})

		It("uses archive downloader", func() {
			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(download.For(env).Download(p, res, "result", env))
			Expect(accepted).To(BeTrue())
			Expect(path).To(Equal("result"))
			Expect(buf.String()).To(StringEqualTrimmedWithContext(`
result: 2 file(s) with 25 byte(s) written
`))

			data := Must(vfs.ReadFile(env, "result/testfile"))
			Expect(string(data)).To(StringEqualWithContext("testdata\n"))
			data = Must(vfs.ReadFile(env, "result/dir/nestedfile"))
			Expect(string(data)).To(StringEqualWithContext("other test data\n"))
		})
	})
})
