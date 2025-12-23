//go:build integration

package artifactset_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
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
		resourcesOciDir string
		resourcesOcmDir string
		env             *TestEnv
	)

	BeforeAll(func() {

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

	It("creates CTF using stable OCM release", func() {
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

		// Use the current OCM version to create the CTF
		GinkgoWriter.Printf("Creating CTF using current OCM version\n")

		buf := bytes.NewBuffer(nil)
		err = env.CatchOutput(buf).Execute(
			"add", "componentversions",
			"--create",
			"--file", sourceCTF,
			constructorFile,
		)
		GinkgoWriter.Printf("OCM output: %s\n", buf.String())
		Expect(err).To(Succeed(), "OCM MUST create CTF: %s", buf.String())
	})

	It("transfers CTF to new CTF with --copy-resources", func() {
		targetCTF = filepath.Join(tempDir, "ctf-target")
		buf := bytes.NewBuffer(nil)
		GinkgoWriter.Printf(" #### transfer componentversions " + sourceCTF + " " + targetCTF + "--copy-resources")
		Expect(env.CatchOutput(buf).Execute(
			"transfer", "componentversions", sourceCTF, targetCTF, "--copy-resources")).To(Succeed())
		GinkgoWriter.Printf("Transfer output: %s\n", buf.String())
	})

	It("verifies components and resources in target CTF", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "componentversions", targetCTF)).To(Succeed())
		GinkgoWriter.Printf("Components: %s\n", buf.String())
		Expect(buf.String()).To(ContainSubstring(componentName))

		// List resources
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute(
			"get", "resources",
			targetCTF+"//"+componentName+":"+componentVersion,
		)).To(Succeed())
		GinkgoWriter.Printf("Resources: %s\n", buf.String())
		Expect(buf.String()).To(ContainSubstring(resourceName))
	})

	It("downloads resource from target CTF with --oci-layout", func() {
		resourcesOciDir = filepath.Join(tempDir, "resource-oci-layout")

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"download", "resources",
			"--oci-layout",
			"-O", resourcesOciDir,
			targetCTF+"//"+componentName+":"+componentVersion,
			resourceName,
		)).To(Succeed())
		Expect(verifyOCILayoutStructure(resourcesOciDir)).To(Succeed())
		store, err := oci.New(resourcesOciDir)
		Expect(err).To(Succeed(), "ORAS failed to open OCI layout: %w", err)

		srcDesc, err := store.Resolve(context.Background(), resourceVersion)
		Expect(err).To(Succeed(), "resource MUST be OCI compliant")
		GinkgoWriter.Printf("Successfully verified OCI layout with ORAS: digest=%s\n", srcDesc.Digest)

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
		GinkgoWriter.Printf("Resource download output: %s\n", buf.String())
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

	GinkgoWriter.Printf("OCI layout structure verified: oci-layout, index.json, blobs/\n")
	return nil
}
