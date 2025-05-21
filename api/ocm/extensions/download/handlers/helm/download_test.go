package helm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "ocm.software/ocm/api/helper/builder"

	"github.com/mandelsoft/filepath/pkg/filepath"
	"github.com/mandelsoft/vfs/pkg/vfs"

	envhelper "ocm.software/ocm/api/helper/env"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/ocm/extensions/download/handlers/helm"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/selectors"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/tarutils"
)

const (
	CTFPath     = "/ctf"
	Component   = "ocm.software/test-component"
	Version     = "v1.0.0"
	OCIResource = "helm"

	ArtifactType       = "NotHelmChart"
	SpecialOCIResource = "specialhelm"
	UnusualOCIResource = "unusualhelm"
	LegacyOCIResource  = "legacyhelm"
)

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(envhelper.TestData("./testdata"))

		env.OCMCommonTransport(CTFPath, accessio.FormatDirectory, func() {
			env.Component(Component, func() {
				env.Version(Version, func() {
					env.Resource(OCIResource, Version, resourcetypes.HELM_CHART, v1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(artdesc.MediaTypeImageManifest), filepath.Join("/testdata/test-chart-oci-artifact.tgz"))
					})
					env.Resource(SpecialOCIResource, Version, ArtifactType, v1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(artdesc.MediaTypeImageManifest), filepath.Join("/testdata/test-chart-oci-artifact.tgz"))
					})
					env.Resource(UnusualOCIResource, Version, resourcetypes.HELM_CHART, v1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(artdesc.MediaTypeImageManifest), filepath.Join("/testdata/unusual-ordered-helm-chart.tgz"))
					})
					env.Resource(LegacyOCIResource, Version, resourcetypes.HELM_CHART, v1.LocalRelation, func() {
						env.BlobFromFile(artifactset.MediaType(artdesc.MediaTypeImageManifest), filepath.Join("/testdata/legacy-pre-hip-helm-chart.tgz"))
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads helm chart from oci artifact by default", func() {
		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTFPath, 0o777, env))
		cv := Must(repo.LookupComponentVersion(Component, Version))
		path := Must(download.DownloadResource(env.OCMContext(), Must(cv.SelectResources(selectors.Identity(v1.Identity{"name": OCIResource})))[0], "/resource", download.WithFileSystem(env.FileSystem())))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(env.FileSystem(), path, env.FileSystem()))
		Expect(Must(vfs.DirExists(env.FileSystem(), "/test-chart"))).To(BeTrue())
	})
	It("register helm download handler by name for special artifact type", func() {
		MustBeSuccessful(download.RegisterHandlerByName(env, helm.PATH, nil, download.ForArtifactType(ArtifactType)))

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTFPath, 0o777, env))
		cv := Must(repo.LookupComponentVersion(Component, Version))
		path := Must(download.DownloadResource(env.OCMContext(), Must(cv.SelectResources(selectors.Identity(v1.Identity{"name": SpecialOCIResource})))[0], "/resource", download.WithFileSystem(env.FileSystem())))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(env.FileSystem(), path, env.FileSystem()))
		Expect(Must(vfs.DirExists(env.FileSystem(), "/test-chart"))).To(BeTrue())
	})

	It("successfully download unusual artifacts with non-defacto helm chart order", func() {
		MustBeSuccessful(download.RegisterHandlerByName(env, helm.PATH, nil, download.ForArtifactType(ArtifactType)))

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTFPath, 0o777, env))
		cv := Must(repo.LookupComponentVersion(Component, Version))
		path := Must(download.DownloadResource(env.OCMContext(), Must(cv.SelectResources(selectors.Identity(v1.Identity{"name": UnusualOCIResource})))[0], "/resource", download.WithFileSystem(env.FileSystem())))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(env.FileSystem(), path, env.FileSystem()))
		Expect(Must(vfs.FileExists(env.FileSystem(), "/postgresql/Chart.yaml"))).To(BeTrue())
	})

	It("successfully download artifacts with helm chart content with legacy content type", func() {
		MustBeSuccessful(download.RegisterHandlerByName(env, helm.PATH, nil, download.ForArtifactType(ArtifactType)))

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTFPath, 0o777, env))
		cv := Must(repo.LookupComponentVersion(Component, Version))
		path := Must(download.DownloadResource(env.OCMContext(), Must(cv.SelectResources(selectors.Identity(v1.Identity{"name": LegacyOCIResource})))[0], "/resource", download.WithFileSystem(env.FileSystem())))
		MustBeSuccessful(tarutils.ExtractArchiveToFs(env.FileSystem(), path, env.FileSystem()))
		Expect(Must(vfs.FileExists(env.FileSystem(), "/ingress-nginx/Chart.yaml"))).To(BeTrue())
	})
})
