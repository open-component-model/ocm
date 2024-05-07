package blueprint_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/projectionfs"

	"github.com/open-component-model/ocm/pkg/common"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/download/handlers/blueprint"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	tenv "github.com/open-component-model/ocm/pkg/env"
	"github.com/open-component-model/ocm/pkg/utils/tarutils"
)

var _ = Describe("blueprint downloader registration", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(tenv.TestData())

		MustBeSuccessful(tarutils.CreateTarFromFs(Must(projectionfs.New(env, TESTDATA_PATH)), ARCHIVE_PATH, tarutils.Gzip, env))

		env.OCICommonTransport(OCI, accessio.FormatDirectory, func() {
			env.Namespace(OCINAMESPACE, func() {
				env.Manifest(OCIVERSION, func() {
					env.Config(func() {
						env.BlobStringData(MIMETYPE, "{}")
					})
					env.Layer(func() {
						env.BlobFromFile(MIMETYPE, ARCHIVE_PATH)
					})
				})
			})
		})

		testhelper.FakeOCIRepo(env, OCI, OCIHOST)
		env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENT, VERSION, func() {
				env.Resource(OCI_ARTIFACT_NAME, ARTIFACT_VERSION, ARTIFACT_TYPE, v1.ExternalRelation, func() {
					env.Access(ociartifact.New(OCIHOST + ".alias/" + OCINAMESPACE + ":" + OCIVERSION))
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("register and use blueprint downloader for artifact type \"testartifacttype\"", func() {
		// As the handler is not registered for the artifact type "testartifacttype" per default (thus, in the
		// init-function of handler.go), this test fails if the registration does not work.
		Expect(download.For(env).RegisterByName(blueprint.PATH, env.OCMContext(), &blueprint.Config{[]string{MIMETYPE}}, download.ForArtifactType(ARTIFACT_TYPE))).To(BeTrue())

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)
		racc := Must(cv.GetResourceByIndex(0))

		p, buf := common.NewBufferedPrinter()
		ok, path := Must2(download.For(env).Download(p, racc, DOWNLOAD_PATH, env))
		Expect(ok).To(BeTrue())
		Expect(path).To(Equal(DOWNLOAD_PATH))
		Expect(env.FileExists(DOWNLOAD_PATH + "/blueprint.yaml")).To(BeTrue())
		Expect(env.FileExists(DOWNLOAD_PATH + "/test/README.md")).To(BeTrue())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(DOWNLOAD_PATH + ": 2 file(s) with 390 byte(s) written"))
	})
})
