package download_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"

	envhelper "ocm.software/ocm/api/helper/env"
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

	// Issue 1668: Downloading artifact leads to broken OCI image layout
	// https://github.com/open-component-model/ocm/issues/1668
	//
	// When running `ocm download artifacts --oci-layout` the directory structure
	// should adhere to the OpenContainer Image Layout Spec.
	// Spec: https://specs.opencontainers.org/image-spec/image-layout/
	//
	// This test suite verifies compliance with each requirement from the spec.
	// Download is executed once, then ORAS library is used to verify all requirements.
	Describe("OCI Image Layout Spec compliance with --oci-layout flag", Ordered, func() {
		var (
			tempDir string
			testEnv *TestEnv
			store   *oci.Store
			ctx     context.Context
		)

		BeforeAll(func() {
			ctx = context.Background()

			var err error
			tempDir, err = os.MkdirTemp("", "oci-test-*")
			Expect(err).To(Succeed())

			fs, err := projectionfs.New(osfs.New(), tempDir)
			Expect(err).To(Succeed())

			testEnv = NewTestEnv(envhelper.FileSystem(fs))

			buf := bytes.NewBuffer(nil)
			Expect(testEnv.CatchOutput(buf).Execute("download", "artifact", "--oci-layout", "-O", "/out", "alpine:latest")).To(Succeed())

			// Open with ORAS - validates oci-layout and index.json
			store, err = oci.New(filepath.Join(tempDir, "out"))
			Expect(err).To(Succeed(), "ORAS MUST be able to open OCI-compliant layout")

		})

		AfterAll(func() {
			Expect(testEnv.Cleanup()).To(Succeed())
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		It("without --oci-layout flag ORAS MUST fail to resolve or fetch", func() {
			// Download without --oci-layout
			outDir := filepath.Join(tempDir, "no-oci-layout")

			buf := bytes.NewBuffer(nil)
			Expect(testEnv.CatchOutput(buf).Execute("download", "artifact", "-O", outDir, "alpine:latest")).To(Succeed())

			// ORAS can open the layout (index.json exists)
			noOciStore, err := oci.New(outDir)
			Expect(err).To(Succeed(), "ORAS opens the layout")

			// Either resolve fails (tag not stored properly) or fetch fails (wrong blob paths)
			desc, resolveErr := noOciStore.Resolve(ctx, "latest")
			Expect(resolveErr).To(Succeed())
			_, fetchErr := content.FetchAll(ctx, noOciStore, desc)
			Expect(fetchErr).NotTo(Succeed(), "ORAS MUST fail to fetch blobs from non-OCI-compliant layout")
		})

		It("manifest MUST be resolvable and fetchable via ORAS", func() {
			manifestDesc, err := store.Resolve(ctx, "latest")
			Expect(err).To(Succeed(), "ORAS MUST resolve tag")

			_, err = content.FetchAll(ctx, store, manifestDesc)
			Expect(err).To(Succeed(), "Manifest MUST be fetchable")
		})

	})

})
