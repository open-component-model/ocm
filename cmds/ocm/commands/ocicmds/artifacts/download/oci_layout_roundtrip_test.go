//go:build integration

package download_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/vfs/pkg/osfs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"github.com/testcontainers/testcontainers-go/wait"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/oci"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	envhelper "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

// This test verifies the complete roundtrip workflow:
// 1. Create a component version with an OCI image resource
// 2. Transfer it to a remote OCI registry (GHCR)
// 3. Download the image using OCM's --oci-layout flag (creates OCI Image Layout)
// 4. Verify the OCI layout is valid and can be read by ORAS
// 5. Copy the OCI layout back to the registry with a new tag using ORAS
// 6. Run the copied image with Docker to confirm it works end-to-end
//
// This confirms that OCM's --oci-layout output is OCI-compliant and interoperable.
var _ = Describe("OCM to ORAS to Docker roundtrip", Ordered, func() {
	const (
		componentName    = "example.com/hello"
		componentVersion = "1.0.0"
		resourceName     = "hello-image"
		resourceVersion  = "1.0.0"
		imageName        = "hello-world"
		imageTag         = "linux"
		copiedImageTag   = "copied-from-oci-layout"
	)

	const distributionRegistryImage = "registry:2.8.3"

	var (
		tempDir           string
		ctfDir            string
		ociDir            string
		env               *TestEnv
		ctx               context.Context
		store             *oci.Store
		srcDesc           ocispec.Descriptor
		copiedImageRef    string
		authClient        *auth.Client
		registryURL       string // with scheme for OCM CLI (e.g., http://localhost:5000)
		registryHost      string // without scheme for ORAS (e.g., localhost:5000)
		registryContainer *registry.RegistryContainer
		log               logr.Logger
	)

	BeforeAll(func() {
		log = GinkgoLogr
		ctx = context.Background()

		// Start containerized registry using testcontainers with wait strategy
		log.Info("Launching test registry", "image", distributionRegistryImage)
		var err error
		registryContainer, err = registry.Run(ctx, distributionRegistryImage,
			testcontainers.WithWaitStrategy(wait.ForHTTP("/v2/").WithPort("5000/tcp")),
		)
		Expect(err).To(Succeed(), "Failed to start registry container")

		registryHost, err = registryContainer.HostAddress(ctx)
		Expect(err).To(Succeed(), "Failed to get registry host address")

		// Use plain HTTP for local testcontainer registry
		registryURL = "http://" + registryHost
		log.Info("Test registry ready", "url", registryURL, "host", registryHost)

		tempDir, err = os.MkdirTemp("", "ocm-integration-*")
		Expect(err).To(Succeed())

		env = NewTestEnv(envhelper.FileSystem(osfs.New()))

		// Create unauthenticated client for ORAS operations (local registry doesn't need auth)
		authClient = &auth.Client{
			Client: retry.DefaultClient,
			Cache:  auth.DefaultCache,
		}
	})

	AfterAll(func() {
		log.Info("Cleaning up registry packages", "host", registryHost)

		// List of repositories to clean up (use registryHost for ORAS)
		reposToClean := []string{
			registryHost + "/" + componentName,     // component version
			registryHost + "/" + imageName,         // copied image
			registryHost + "/library/" + imageName, // transferred image (OCM uses library/ prefix)
		}

		for _, repoPath := range reposToClean {
			repo, err := remote.NewRepository(repoPath)
			if err != nil {
				continue
			}
			repo.Client = authClient

			// Try to delete all known tags
			tagsToDelete := []string{componentVersion, imageTag, copiedImageTag}
			for _, tag := range tagsToDelete {
				desc, err := repo.Resolve(ctx, tag)
				if err != nil {
					continue
				}
				if err := repo.Delete(ctx, desc); err != nil {
					log.Info("Warning: failed to delete", "repo", repoPath, "tag", tag, "error", err)
				} else {
					log.Info("Deleted", "repo", repoPath, "tag", tag)
				}
			}
		}

		if env != nil {
			env.Cleanup()
		}
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}

		// Terminate the registry container
		if registryContainer != nil {
			if err := testcontainers.TerminateContainer(registryContainer); err != nil {
				log.Info("Warning: failed to terminate registry container", "error", err)
			}
		}
	})

	// Step 1: Create a component version with an OCI image resource.
	// This simulates a user packaging their application with OCM.
	It("creates component version from constructor", func() {
		ctfDir = filepath.Join(tempDir, "ctf")
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
          imageReference: ` + imageName + `:` + imageTag + `
`
		err := os.WriteFile(constructorFile, []byte(constructorContent), 0644)
		Expect(err).To(Succeed(), "MUST create constructor file")

		// Create component version
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"add", "componentversions",
			"--create",
			"--file", ctfDir,
			constructorFile,
		)).To(Succeed())
		log.Info("Create output", "output", buf.String())
	})

	// Step 2: Verify the component version was created correctly.
	// Lists the component and its resources to confirm the structure.
	It("lists components and resources", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("get", "componentversions", ctfDir)).To(Succeed())
		log.Info("Components", "output", buf.String())
		Expect(buf.String()).To(ContainSubstring(componentName))

		// List resources
		buf.Reset()
		Expect(env.CatchOutput(buf).Execute(
			"get", "resources",
			ctfDir+"//"+componentName+":"+componentVersion,
		)).To(Succeed())
		log.Info("Resources", "output", buf.String())
		Expect(buf.String()).To(ContainSubstring(resourceName))
	})

	// Step 3: Transfer the component version to a remote OCI registry.
	// Uses --copy-resources to also push the referenced OCI image.
	It("transfers component version to a remote registry", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"transfer", "componentversions",
			"--copy-resources",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)).To(Succeed())
		log.Info("Transfer output", "output", buf.String())
	})

	// Step 4: Verify the component version exists in the remote registry.
	// Confirms the transfer was successful.
	It("lists resources from remote registry", func() {
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"get", "resources",
			registryURL+"//"+componentName+":"+componentVersion,
		)).To(Succeed())
		log.Info("Remote resources", "output", buf.String())
		Expect(buf.String()).To(ContainSubstring(resourceName))
	})

	// Step 5: Download an OCI artifact using --oci-layout flag.
	// This creates an OCI Image Layout directory structure that should be
	// compliant with the OCI Image Layout Specification.
	It("downloads OCI image with --oci-layout", func() {
		ociDir = filepath.Join(tempDir, "oci-download")

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"download", "artifact",
			"--oci-layout",
			"-O", ociDir,
			imageName+":"+imageTag,
		)).To(Succeed())
		log.Info("Download output", "output", buf.String())
	})

	// Step 6: Verify the OCI layout is valid using ORAS.
	// If ORAS can open and resolve the layout, it confirms OCI compliance.
	It("opens downloaded OCI layout with ORAS", func() {
		var err error
		store, err = oci.New(ociDir)
		Expect(err).To(Succeed(), "ORAS MUST open OCI layout")

		// Resolve tag and enumerate content (layers/config)
		srcDesc, err = store.Resolve(ctx, imageTag)
		Expect(err).To(Succeed(), "ORAS MUST resolve tag")

		successors, err := content.Successors(ctx, store, srcDesc)
		Expect(err).To(Succeed(), "ORAS MUST get successors")
		Expect(len(successors)).To(BeNumerically(">", 0))
		log.Info("Found successors", "count", len(successors))
	})

	// Step 7: Copy the OCI layout back to the registry with a new tag.
	// This proves the downloaded OCI layout can be republished using standard tools.
	It("copies OCI layout to registry with new tag", func() {
		copiedImageRef = registryHost + "/" + imageName + ":" + copiedImageTag
		dst, err := remote.NewRepository(registryHost + "/" + imageName)
		Expect(err).To(Succeed())
		dst.Client = authClient
		dst.PlainHTTP = true // testcontainers registry uses plain HTTP
		desc, err := oras.Copy(ctx, store, imageTag, dst, copiedImageTag, oras.DefaultCopyOptions)
		Expect(err).To(Succeed())
		log.Info("Copied", "ref", copiedImageRef, "digest", desc.Digest)
	})

	// Step 8: Verify the copied image exists in the registry.
	// Confirms the ORAS copy operation was successful.
	It("verifies copied image in registry", func() {
		repo, err := remote.NewRepository(registryHost + "/" + imageName)
		Expect(err).To(Succeed())
		repo.Client = authClient
		repo.PlainHTTP = true // testcontainers registry uses plain HTTP
		desc, err := repo.Resolve(ctx, copiedImageTag)
		Expect(err).To(Succeed())
		log.Info("Verified image in registry", "ref", copiedImageRef, "digest", desc.Digest)
	})

	// Step 9: Run the copied image with Docker.
	// This is the final proof that the entire roundtrip works:
	// OCM create -> transfer -> download (--oci-layout) -> ORAS copy -> Docker run
	It("runs copied image from registry with docker", func() {
		log.Info("Running image", "ref", copiedImageRef)
		cmd := exec.Command("docker", "run", "--rm", copiedImageRef)
		out, err := cmd.CombinedOutput()
		Expect(err).To(Succeed(), "Docker run failed: %s", string(out))
		Expect(string(out)).To(ContainSubstring("Hello from Docker"))
		log.Info("Docker output", "output", string(out))
	})
})
