package genericocireg_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/version"
)

const (
	ARCH          = "/tmp/ctf"
	PROVIDER      = "mandelsoft"
	VERSION       = "v1"
	TEST_CONTENT1 = "this is a test"
	TEST_CONTENT2 = "this is another test"
)

var _ = Describe("Transfer handler", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("test1", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, TEST_CONTENT1)
					})
					env.Resource("test2", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, TEST_CONTENT1)
					})
					env.Resource("test3", "", resourcetypes.PLAIN_TEXT, metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, TEST_CONTENT2)
					})
					env.Source("test4", "", resourcetypes.PLAIN_TEXT, func() {
						env.BlobStringData(mime.MIME_TEXT, TEST_CONTENT2)
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("check expected oci layer annotation", func() {
		arch := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(arch)
		c := Must(arch.LookupNamespace("component-descriptors/" + COMPONENT))
		defer Close(c)
		v := Must(c.GetArtifact(VERSION))
		defer Close(v)
		Expect(v.GetDescriptor()).To(YAMLEqual(`
  mediaType: application/vnd.oci.image.manifest.v1+json
  schemaVersion: 2
  annotations:
    ` + genericocireg.OCM_COMPONENTVERSION + `: github.com/mandelsoft/ocm:v1
    ` + genericocireg.OCM_CREATOR + `: OCM Go Library ` + version.Current() + `
  config:
    digest: sha256:edf034a303e8cc7e5a05c522bb5fc74a09a00ed3aca390ffafba1020c97470cc
    mediaType: application/vnd.ocm.software.component.config.v1+json
    size: 201
  layers:
  - digest: sha256:43d7113c5e6dd84c617477c5713368cffb86b059808df176bbf3e02849ea6b3e
    mediaType: application/vnd.ocm.software.component-descriptor.v2+yaml+tar
    size: 3584
  - annotations:
      ` + genericocireg.OCM_ARTIFACT + `: '[{"kind":"resource","identity":{"name":"test1"}},{"kind":"resource","identity":{"name":"test2"}}]'
    digest: sha256:2e99758548972a8e8822ad47fa1017ff72f06f3ff6a016851f45c398732bc50c
    mediaType: text/plain
    size: 14
  - annotations:
      ` + genericocireg.OCM_ARTIFACT + `: '[{"kind":"resource","identity":{"name":"test3"}},{"kind":"source","identity":{"name":"test4"}}]'
    digest: sha256:f69bff44070ba35d7169196ba0095425979d96346a31486b507b4a3f77bd356d
    mediaType: text/plain
    size: 20
`))
	})
})
