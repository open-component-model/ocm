//go:build integration

package artifactset_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content/oci"

	envhelper "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	componentName    = "example.com/hello"
	componentVersion = "1.0.0"
	resourceName     = "hello-image"
	resourceVersion  = "1.0.0"
	imageReference   = "hello-world:linux"
)

// This test verifies the CTF-based workflow with --oci-layout flag:
//  1. Create a CTF archive with an OCI image resource
//  2. Transfer CTF to new CTF with --copy-resources
//  3. Verify components and resources in target CTF
//  4. Download resource with --oci-layout flag:
//     - Creates OCI Image Layout directory (index.json, oci-layout, blobs/sha256/...)
//     - Verifies layout structure is OCI-compliant
//     - Resolves artifact by resource version using ORAS
//  5. Download resource without --oci-layout:
//     - Creates OCM artifact set format (not OCI-compliant)
//     - Verifies layout structure check fails
var _ = Describe("CTF to CTF-with-resource to OCI roundtrip", Ordered, func() {
	var (
		tempDir         string
		sourceCTF       string
		targetCTF       string
		resourcesOciTgz string
		resourcesOcmDir string
		imageTag        string
		env             *TestEnv
		log             logr.Logger
	)

	BeforeAll(func() {
		log = GinkgoLogr

		var err error
		tempDir, err = os.MkdirTemp("", "ocm-ctf-oci-layout-*")
		Expect(err).To(Succeed())

		env = NewTestEnv(envhelper.FileSystem(osfs.New()))
	})

	AfterAll(func() {
		if env != nil {
			env.Cleanup()
		}
	})

	It("creates CTF ", func() {
		sourceCTF = filepath.Join(tempDir, "ctf-source")
		constructorFile := filepath.Join(tempDir, "component-constructor.yaml")
		constructorContent := `components:
  - name: ` + componentName + `
    version: ` + componentVersion + `
    provider:
      name: example.com
    resources:
      - name: ` + resourceName + `
        type: ociImage
        version: ` + resourceVersion + `
        relation: external
        access:
          type: ociArtifact
          imageReference: ` + imageReference + `
`
		err := os.WriteFile(constructorFile, []byte(constructorContent), 0644)
		Expect(err).To(Succeed(), "MUST create constructor file")

		// Create CTF directory
		err = os.MkdirAll(sourceCTF, 0755)
		Expect(err).To(Succeed(), "MUST create CTF directory")
		log.Info("Creating CTF using current OCM version")

		buf := bytes.NewBuffer(nil)
		err = env.CatchOutput(buf).Execute(
			"add", "componentversions",
			"--create",
			"--file", sourceCTF,
			constructorFile,
		)
		log.Info("OCM output", "output", buf.String())
		Expect(err).To(Succeed(), "OCM MUST create CTF: %s", buf.String())
	})

	It("transfers CTF to new CTF with --copy-resources", func() {
		targetCTF = filepath.Join(tempDir, "ctf-target")
		buf := bytes.NewBuffer(nil)
		log.Info("transfer componentversions", "source", sourceCTF, "target", targetCTF)
		Expect(env.CatchOutput(buf).Execute(
			"transfer", "componentversions", sourceCTF, targetCTF, "--copy-resources")).To(Succeed())
		log.Info("Transfer output", "output", buf.String())
	})

	It("verifies components and resources in target CTF", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "componentversions", targetCTF)).To(Succeed())
		log.Info("Components", "output", buf.String())
		Expect(buf.String()).To(ContainSubstring(componentName))

		// List resources
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute(
			"get", "resources",
			targetCTF+"//"+componentName+":"+componentVersion,
		)).To(Succeed())
		log.Info("Resources", "output", buf.String())
		Expect(buf.String()).To(ContainSubstring(resourceName))
	})

	It("downloads resource as OCI tgz with --oci-layout", func() {
		resourcesOciTgz = filepath.Join(tempDir, "resource-oci-layout.tgz")

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"download", "resources", "--oci-layout",
			"-O", resourcesOciTgz,
			targetCTF+"//"+componentName+":"+componentVersion, resourceName,
		)).To(Succeed())
		log.Info("Downloaded OCI tgz", "path", resourcesOciTgz)
	})

	It("verifies with oras-go library", func() {
		ctx := context.Background()
		// Gunzip for oras-go (doesn't support gzip directly)
		tarPath := resourcesOciTgz[:len(resourcesOciTgz)-3] + "tar"
		gunzipCmd := exec.Command("sh", "-c", "gunzip -c "+resourcesOciTgz+" > "+tarPath)
		out, err := gunzipCmd.CombinedOutput()
		Expect(err).To(Succeed(), "gunzip failed: %s", string(out))

		// Open OCI layout from tar using oras-go
		store, err := oci.NewFromTar(ctx, tarPath)
		Expect(err).To(Succeed(), "oras failed to open tar as OCI layout")

		// Resolve by resource version tag
		desc, err := store.Resolve(ctx, resourceVersion)
		Expect(err).To(Succeed(), "oras failed to resolve by resource version tag")
		Expect(desc.MediaType).To(Equal(ociv1.MediaTypeImageIndex))
		log.Info("ORAS verified OCI layout", "tag", resourceVersion, "digest", desc.Digest.String())

		// Fetch index
		reader, err := store.Fetch(ctx, desc)
		Expect(err).To(Succeed(), "failed to fetch manifest")
		manifestData, err := io.ReadAll(reader)
		Expect(err).To(Succeed(), "failed to read manifest")
		Expect(reader.Close()).To(Succeed(), "failed to close reader")
		var index ociv1.Index
		Expect(json.Unmarshal(manifestData, &index)).To(Succeed())

		// Fetch manifest
		var manifest ociv1.Manifest
		reader, err = store.Fetch(ctx, index.Manifests[0])
		Expect(err).To(Succeed(), "failed to fetch manifest")
		manifestData, err = io.ReadAll(reader)
		Expect(err).To(Succeed(), "failed to read manifest")
		Expect(reader.Close()).To(Succeed(), "failed to close reader")
		Expect(json.Unmarshal(manifestData, &manifest)).To(Succeed())
		Expect(manifest.Layers).To(HaveLen(1))

		// Verify config
		configReader, err := store.Fetch(ctx, manifest.Config)
		Expect(err).To(Succeed(), "failed to fetch config")
		configData, err := io.ReadAll(configReader)
		Expect(err).To(Succeed(), "failed to read config")
		Expect(configReader.Close()).To(Succeed(), "failed to close reader")

		var config ociv1.Image
		Expect(json.Unmarshal(configData, &config)).To(Succeed())
		Expect(config.Config.Cmd).ToNot(BeEmpty(), "image config should have Cmd")
		log.Info("Verified image config", "cmd", config.Config.Cmd)
	})

	It("copies OCI archive to Docker with skopeo", func() {
		// Use skopeo to copy from OCI archive (tgz) to docker daemon
		imageTag = "ocm-test-hello:" + resourceVersion
		cmd := exec.Command("skopeo", "copy",
			"oci-archive:"+resourcesOciTgz+":"+resourceVersion,
			"docker-daemon:"+imageTag,
			"--override-os=linux")
		out, err := cmd.CombinedOutput()
		Expect(err).To(Succeed(), "skopeo copy failed: %s", string(out))
		log.Info("Skopeo copy output", "output", string(out))
	})

	It("runs image copied by skopeo", func() {
		log.Info("Running image", "tag", imageTag)

		cmd := exec.Command("docker", "run", "--rm", imageTag)
		out, err := cmd.CombinedOutput()
		Expect(err).To(Succeed(), "docker run failed: %s", string(out))
		Expect(string(out)).To(ContainSubstring("Hello from Docker"))

		// Cleanup
		_ = exec.Command("docker", "rmi", imageTag).Run()
	})

	It("downloads resource from target CTF without --oci-layout", func() {
		resourcesOcmDir = filepath.Join(tempDir, "resource-ocm-layout")

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"download", "resources",
			"-O", resourcesOcmDir,
			targetCTF+"//"+componentName+":"+componentVersion,
			resourceName,
		)).To(Succeed())
		log.Info("Resource download output", "output", buf.String())

		// Verify oras cannot open OCM format as OCI layout
		_, err := oci.New(resourcesOcmDir)
		Expect(err).ToNot(Succeed(), "oras should fail to open non-OCI layout")
		log.Info("ORAS correctly rejected non-OCI layout", "error", err)
	})
})
