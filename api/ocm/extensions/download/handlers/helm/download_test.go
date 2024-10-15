package helm_test

import (
	"github.com/mandelsoft/filepath/pkg/filepath"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	envhelper "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/ocm/extensions/download"
	registry "ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/helm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/tarutils"

	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/utils/accessio"
)

const (
	CTFPath     = "/ctf"
	Component   = "ocm.software/test-component"
	Version     = "v1.0.0"
	OCIResource = "helm"
)

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(envhelper.TestData("./testdata"))

		env.OCMCommonTransport(CTFPath, accessio.FormatDirectory, func() {
			env.Component(Component, func() {
				env.Version(Version, func() {
					env.Resource(OCIResource, Version, resourcetypes.HELM_CHART, v1.LocalRelation, func() {
						env.BlobFromFile(artdesc.MediaTypeImageManifest, filepath.Join("/testdata/test-chart-oci-artifact.tgz"))
					})
				})
			})
		})

	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads helm chart from oci artifact", func() {
		download.For(env.OCMContext()).Register(helm.New(), registry.ForArtifactType(helm.TYPE))
		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTFPath, 0o777, env))
		cv := Must(repo.LookupComponentVersion(Component, Version))
		path := Must(download.DownloadResource(env.OCMContext(), Must(cv.SelectResources(selectors.Identity(v1.Identity{"name": OCIResource})))[0], "/resource", download.WithFileSystem(env.FileSystem())))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(env.FileSystem(), path, env.FileSystem()))
		Expect(Must(vfs.DirExists(env.FileSystem(), "/test-chart"))).To(BeTrue())
	})
})
