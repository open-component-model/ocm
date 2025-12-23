//go:build integration

package artifactset_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		resolvedDigest  string
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

	It("verifies ref.name with oras-go library", func() {
		// Gunzip for oras-go (doesn't support gzip directly)
		tarPath := resourcesOciTgz[:len(resourcesOciTgz)-3] + "tar"
		cmd := exec.Command("sh", "-c", "gunzip -c "+resourcesOciTgz+" > "+tarPath)
		out, err := cmd.CombinedOutput()
		Expect(err).To(Succeed(), "gunzip failed: %s", string(out))

		// Open OCI layout from tar using oras-go
		store, err := oci.NewFromTar(context.Background(), tarPath)
		Expect(err).To(Succeed(), "oras failed to open tar")

		// Resolve by resource version tag - this proves ref.name is set correctly
		desc, err := store.Resolve(context.Background(), resourceVersion)
		Expect(err).To(Succeed(), "oras failed to resolve by resource version")
		resolvedDigest = desc.Digest.String()
		log.Info("ORAS resolved manifest by resource version", "version", resourceVersion, "digest", resolvedDigest)
	})

	It("copies OCI archive to Docker with skopeo", func() {
		// Use skopeo to copy from OCI archive (tgz) to docker daemon
		imageTag = "ocm-test-hello:" + resourceVersion
		cmd := exec.Command("skopeo", "copy",
			"oci-archive:"+resourcesOciTgz+":"+resourceVersion,
			"docker-daemon:"+imageTag)
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
		Expect(verifyOCILayoutStructure(resourcesOcmDir)).ToNot(Succeed())
		log.Info("Resource download output", "output", buf.String())
	})
})

// verifyOCILayoutStructure checks that the OCI layout has the expected structure.
// Returns an error if any required file or directory is missing.
func verifyOCILayoutStructure(ociDir string) error {
	// Check oci-layout file exists
	ociLayoutPath := filepath.Join(ociDir, "oci-layout")
	if _, err := os.Stat(ociLayoutPath); err != nil {
		return fmt.Errorf("oci-layout file MUST exist: %w", err)
	}

	// Check index.json exists
	indexPath := filepath.Join(ociDir, "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("index.json MUST exist: %w", err)
	}

	// Check blobs directory exists
	blobsDir := filepath.Join(ociDir, "blobs")
	info, err := os.Stat(blobsDir)
	if err != nil {
		return fmt.Errorf("blobs directory MUST exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("blobs MUST be a directory, got file")
	}

	return nil
}
