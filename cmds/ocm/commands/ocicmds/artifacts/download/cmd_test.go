package download_test

import (
	"bytes"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH    = "/tmp/ctf"
	VERSION = "v1"
	NS      = "mandelsoft/test"
	OUT     = "/tmp/res"
)

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
		env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Namespace(NS, func() {
				env.Manifest(VERSION, func() {
					env.Config(func() {
						env.BlobStringData(mime.MIME_JSON, "{}")
					})
					env.Layer(func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("downloads single artifact from ctf file", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "artifact", "-O", OUT, "--repo", ARCH, NS+grammar.TagSeparator+VERSION)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: downloaded
`))
		Expect(env.DirExists(OUT)).To(BeTrue())
		tags := ""
		if artifactset.IsOCIDefaultFormat() {
			tags = "\"org.opencontainers.image.ref.name\":\"v1\","
		}
		sha := "sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9"
		Expect(env.ReadFile(OUT + "/" + artifactset.DefaultArtifactSetDescriptorFileName)).To(Equal([]byte("{\"schemaVersion\":2,\"mediaType\":\"application/vnd.oci.image.index.v1+json\",\"manifests\":[{\"mediaType\":\"application/vnd.oci.image.manifest.v1+json\",\"digest\":\"sha256:2c3e2c59e0ac9c99864bf0a9f9727c09f21a66080f9f9b03b36a2dad3cce6ff9\",\"size\":342,\"annotations\":{" + tags + "\"software.ocm/tags\":\"v1\"}}],\"annotations\":{\"software.ocm/main\":\"" + sha + "\"}}")))
	})

	It("download single artifact layer from ctf file", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("download", "artifact", "--layers=0", "-O", OUT, "--repo", ARCH, NS+grammar.TagSeparator+VERSION)).To(Succeed())
		Expect(buf.String()).To(StringEqualTrimmedWithContext(
			`
/tmp/res: layer 0: 8 byte(s) downloaded
`))
		Expect(env.ReadFile("/tmp/res")).To(StringEqualWithContext("testdata"))
	})
})
