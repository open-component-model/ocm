package transfer_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"golang.org/x/crypto/bcrypt"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessobj"
)

const distributionRegistryImage = "registry:3.0.0"

type TestLogger struct{}

func (t TestLogger) Printf(format string, v ...interface{}) {
	GinkgoWriter.Printf(format+"\n", v...)
}

func StartDockerContainerRegistry(t FullGinkgoTInterface, container, htpasswd string) string {
	t.Helper()
	// Start containerized registry
	t.Logf("Launching test registry (%s)...", distributionRegistryImage)
	registryContainer, err := registry.Run(context.Background(), distributionRegistryImage,
		WithHtpasswd(htpasswd),
		testcontainers.WithEnv(map[string]string{
			"REGISTRY_VALIDATION_DISABLED": "true",
			"REGISTRY_LOG_LEVEL":           "debug",
		}),
		testcontainers.WithLogger(&TestLogger{}),
		testcontainers.WithName(container),
	)
	r := require.New(t)
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(testcontainers.TerminateContainer(registryContainer))
	})
	t.Logf("Test registry started")

	registryAddress, err := registryContainer.HostAddress(context.Background())
	r.NoError(err)

	return registryAddress
}

func WithHtpasswd(credentials string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		tmpFile, err := os.CreateTemp("", "htpasswd")
		if err != nil {
			tmpFile, err = os.Create(".")
			if err != nil {
				return fmt.Errorf("cannot create the file in the temp dir or in the current dir: %w", err)
			}
		}
		defer tmpFile.Close()

		_, err = tmpFile.WriteString(credentials)
		if err != nil {
			return fmt.Errorf("cannot write the credentials to the file: %w", err)
		}

		return registry.WithHtpasswdFile(tmpFile.Name())(req)
	}
}

func GenerateHtpasswd(t FullGinkgoTInterface, username, password string) string {
	t.Helper()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return fmt.Sprintf("%s:%s", username, hashedPassword)
}

var _ = Describe("Index Transfer", func() {
	var env *TestEnv
	var registryAddress string
	user := "admin"
	pass := "password"

	BeforeEach(func() {
		env = NewTestEnv()
		// Start Local Registry
		registryAddress = StartDockerContainerRegistry(GinkgoT(), "ocm-index-transfer-test-registry-"+time.Now().Format("20060102150405"), GenerateHtpasswd(GinkgoT(), user, pass))
	})

	AfterEach(func() {
		env.Cleanup()
	})

	FIt("transfers an OCI Index artifact as a local blob", func() {
		tempDir := env.FSTempDir()
		ctx := context.Background()
		r := require.New(GinkgoT())

		// 1. Pull the CLI image (Index) to local registry using oras-go
		srcRef := "ghcr.io/open-component-model/cli:main"

		dstRef := fmt.Sprintf("%s/ocm-cli:main", registryAddress)

		// Create a resolver for GHCR
		repoSrc, err := remote.NewRepository(srcRef)
		r.NoError(err)

		desc, err := oras.Resolve(ctx, repoSrc, srcRef, oras.ResolveOptions{})
		r.NoError(err)
		repoSrc.Reference.Reference = desc.Digest.String()
		srcRef = repoSrc.Reference.String()

		repoDst, err := remote.NewRepository(dstRef)
		r.NoError(err)

		repoDst.Client = &auth.Client{
			Client: retry.DefaultClient,
			Credential: auth.StaticCredential(registryAddress, auth.Credential{
				Username: "admin",
				Password: "password",
			}),
		}
		repoDst.PlainHTTP = true

		_, err = oras.Copy(ctx, repoSrc, repoSrc.Reference.Reference, repoDst, repoDst.Reference.Reference, oras.CopyOptions{})
		r.NoError(err, "failed to copy cli image from ghcr to local registry")

		// Define Component Version with dir input using the OCI Layout directory
		componentName := "ocm.software/test-component"
		componentVersion := "v1.0.0"
		resourceName := "cli-image"

		constructorContent := fmt.Sprintf(`
components:
- name: %s
  version: %s
  provider:
    name: ocm.software
  resources:
  - name: %s
    version: %s
    type: ociImage
    relation: external
    copyPolicy: byValue
    access:
      type: ociArtifact
      imageReference: "%s"
`, componentName, componentVersion, resourceName, componentVersion, "http://"+dstRef)

		constructorPath := filepath.Join(tempDir, "constructor.yaml")
		r.NoError(os.WriteFile(constructorPath, []byte(constructorContent), os.ModePerm))

		// 2.1 Generate .ocmconfig for credentials
		// Extract port from registryAddress (e.g. localhost:55007)
		// We assume localhost for testcontainers with host network
		host, port, err := net.SplitHostPort(registryAddress)
		r.NoError(err)

		ocmConfigContent := fmt.Sprintf(`
type: generic.config.ocm.software/v1
configurations:
- type: credentials.config.ocm.software
  consumers:
    - identity:
        type: OCIRepository
        hostname: %[1]q
        port: %[2]q
        scheme: http
      credentials:
      - type: Credentials/v1
        properties:
          username: admin
          password: password
`, host, port)
		configPath := filepath.Join(tempDir, "ocmconfig.yaml")
		r.NoError(os.WriteFile(configPath, []byte(strings.TrimSpace(ocmConfigContent)), os.ModePerm))

		// 3. Add Component Version directly to the local registry using the NEW CLI (via docker run)
		// We mount the tempDir to /work in the container
		// We use --network host so the container can reach the local test registry found on localhost

		dockerArgs := []string{
			"run", "--rm",
			"--network", "host",
			"-v", fmt.Sprintf("%s:/work", tempDir),
			"ghcr.io/open-component-model/cli:0.0.0-local.dev",
			"add", "component-version",
			"--repository", "http://" + registryAddress,
			"--config", "/work/ocmconfig.yaml",
			"--constructor", "/work/constructor.yaml",
			"--loglevel", "debug",
		}

		verifyCmd := exec.Command("docker", dockerArgs...)
		out, err := verifyCmd.CombinedOutput()
		r.NoError(err, fmt.Sprintf("failed to run new CLI via docker: %s", string(out)))

		dockerArgs = []string{
			"run", "--rm",
			"--network", "host",
			"-v", fmt.Sprintf("%s:/work", tempDir),
			"ghcr.io/open-component-model/cli:0.0.0-local.dev",
			"get", "component-version",
			"--config", "/work/ocmconfig.yaml",
			"--loglevel", "error",
			"-oyaml",
			fmt.Sprintf("%s//%s:%s", "http://"+registryAddress, componentName, componentVersion),
		}
		verifyCmd = exec.Command("docker", dockerArgs...)
		out, err = verifyCmd.CombinedOutput()
		r.NoError(err, fmt.Sprintf("failed to verify component version in local registry via docker: %s", string(out)))

		credArgs := []string{
			"--cred", ":type=" + "OCIRegistry",
			"--cred", ":hostname=" + host,
			"--cred", ":port=" + port,
			"--cred", ":scheme=" + "http",
			"--cred", "username=" + user,
			"--cred", "password=" + pass,
		}

		// Ensure the resource can be accessed via ocm CLI
		Expect(env.Execute(append(credArgs,
			"get", "resource",
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress,
			"-ojson",
			componentName+":"+componentVersion, resourceName,
		)...)).To(Succeed())

		// 4a. Download the resource to verify synthesization works on demand
		Expect(env.Execute(append(credArgs,
			"download", "resource",
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress,
			componentName+":"+componentVersion, resourceName,
			"--outfile", filepath.Join(tempDir, "downloaded-cli"),
		)...)).To(Succeed())

		// 4b. Transfer from Registry to CTF with synthesization (ResourcesByValue=true to force localBlob creation)
		targetCTF := filepath.Join(tempDir, "target-ctf")
		Expect(env.Execute(append(credArgs,
			"transfer", "componentversions",
			componentName+":"+componentVersion,
			targetCTF,
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress, // Use the registry as lookup for references too
			"--copy-resources",
		)...)).To(Succeed())

		// 5. Verify that the target CTF contains the resource as a localBlob and can be read
		repo, err := ctf.Open(env, accessobj.ACC_READONLY, targetCTF, 0o700, ctf.FormatDirectory)
		r.NoError(err)
		defer repo.Close()

		cv, err := repo.LookupComponentVersion(componentName, componentVersion)
		r.NoError(err)
		defer cv.Close()

		resources := cv.GetResources()
		var res cpi.ResourceAccess
		for _, resc := range resources {
			if resc.Meta().Name == resourceName {
				res = resc
				break
			}
		}
		r.NotNil(res, "resource not found")

		meth, err := res.AccessMethod()
		r.NoError(err)
		defer meth.Close()

		r.Equal("localBlob", meth.GetKind())

		blob, err := meth.Get()
		r.NoError(err)
		r.NotEmpty(blob)
	})
})
