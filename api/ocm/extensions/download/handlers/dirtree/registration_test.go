package dirtree_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/helper/builder"
	envhelper "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/dirtree"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	common "ocm.software/ocm/api/utils/misc"
	"ocm.software/ocm/api/utils/tarutils"
)

const TEST_ARTIFACT = "testArtifact"

var _ = Describe("artifact management", func() {
	var env *builder.Builder

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
					env.Resource(RESOURCE, VERSION, TEST_ARTIFACT, metav1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(mime.MIME_TGZ_ALT), "archive")
					})
				})
			})
		})

		It("downloads to dir", func() {
			Expect(download.For(env).RegisterByName("ocm/dirtree", env.OCMContext(), &dirtree.Config{AsArchive: false}, download.ForArtifactType(TEST_ARTIFACT))).To(BeTrue())

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

		It("downloads archive to archive", func() {
			Expect(download.For(env).RegisterByName("ocm/dirtree", env.OCMContext(), &dirtree.Config{AsArchive: true}, download.ForArtifactType(TEST_ARTIFACT))).To(BeTrue())

			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(download.For(env).Download(p, res, "target", env))
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

		It("downloads archive to archive using config", func() {
			spec := `
type: downloader.ocm.config.ocm.software
registrations:
- name: ocm/dirtree
  artifactType: ` + TEST_ARTIFACT + `
  config:
    asArchive: true
`
			env.ConfigContext().ApplyData([]byte(spec), nil, "manual")

			repo := Must(ctf.Open(ocm.DefaultContext(), accessobj.ACC_READONLY, "ctf", 0, env))
			defer Close(repo)
			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			defer Close(cv)
			res := Must(cv.GetResource(metav1.NewIdentity(RESOURCE)))

			p, buf := common.NewBufferedPrinter()
			accepted, path := Must2(download.For(env).Download(p, res, "target", env))
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
	})
})
