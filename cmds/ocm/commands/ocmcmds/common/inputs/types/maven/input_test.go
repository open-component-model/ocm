package maven_test

import (
	"crypto"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/tech/maven/maventest"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/tarutils"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	ARCH    = "test.ca"
	VERSION = "v1"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData(), maventest.TestData("/maven/testdata"))

		Expect(env.Execute("create", "ca", "-ft", "directory", "test.de/x", VERSION, "--provider", "mandelsoft", "--file", ARCH)).To(Succeed())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("add maven from file system described by resources.yaml", func() {
		Expect(env.Execute("add", "resources", "--file", ARCH, "/testdata/resources1.yaml")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))
		access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
		Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
		fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
		Expect(fi.Size()).To(Equal(int64(maventest.ARTIFACT_SIZE)))
		li := Must(tarutils.ListArchiveContent(env.Join(ARCH, "blobs", access.LocalReference), env.FileSystem()))
		Expect(li).To(ConsistOf(
			"sdk-modules-bom-5.7.0-random-content.json",
			"sdk-modules-bom-5.7.0-random-content.txt",
			"sdk-modules-bom-5.7.0-sources.jar",
			"sdk-modules-bom-5.7.0.jar",
			"sdk-modules-bom-5.7.0.pom"))
		Expect(cd.Resources[0].Digest.HashAlgorithm).To(Equal(crypto.SHA256.String()))
		Expect(cd.Resources[0].Digest.Value).To(Equal(maventest.ARTIFACT_DIGEST))
	})

	It("add maven from file system described by cli options", func() {
		meta := `
name: testdata
type: mavenPackage
`
		Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "maven",
			"--inputPath", "/maven/testdata/.m2/repository", "--groupId", "com.sap.cloud.sdk", "--artifactId", "sdk-modules-bom",
			"--inputVersion", "5.7.0", "--classifier", "", "--extension", "pom")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))
		access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
		Expect(access.MediaType).To(Equal(mime.MIME_XML))
		fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
		Expect(fi.Size()).To(Equal(int64(7153)))
	})

	It("add maven file from file system described by cli options", func() {
		meta := `
name: testdata
type: mavenPackage
`
		Expect(env.Execute("add", "resources", "--file", ARCH, "--resource", meta, "--inputType", "maven",
			"--inputPath", "/maven/testdata/.m2/repository", "--groupId", "com.sap.cloud.sdk", "--artifactId", "sdk-modules-bom",
			"--inputVersion", "5.7.0", "--extension", "pom")).To(Succeed())
		data, err := env.ReadFile(env.Join(ARCH, comparch.ComponentDescriptorFileName))
		Expect(err).To(Succeed())
		cd, err := compdesc.Decode(data)
		Expect(err).To(Succeed())
		Expect(len(cd.Resources)).To(Equal(1))
		access := Must(env.Context.OCMContext().AccessSpecForSpec(cd.Resources[0].Access)).(*localblob.AccessSpec)
		Expect(access.MediaType).To(Equal(mime.MIME_TGZ))
		fi := Must(env.FileSystem().Stat(env.Join(ARCH, "blobs", access.LocalReference)))
		Expect(fi.Size()).To(Equal(int64(1109)))
	})
})
