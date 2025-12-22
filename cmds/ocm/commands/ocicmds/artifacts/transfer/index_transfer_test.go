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

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	"golang.org/x/crypto/bcrypt"
	"ocm.software/ocm/api/utils/tarutils"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
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

	It("transfers an OCI Index artifact as a local blob", func() {
		ctx := context.Background()
		r := require.New(GinkgoT())
		tempDir := env.FSTempDir()

		user := "admin"
		pass := "password"

		var (
			componentName    = "ocm.software/test-component"
			componentVersion = "v1.0.0"
			resourceName     = "cli-image"
		)

		By("pulling the OCI index image from GHCR and pushing it to the local registry")
		srcRef := "ghcr.io/open-component-model/cli:main"
		dstRef := fmt.Sprintf("%s/ocm-cli:main", registryAddress)

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
				Username: user,
				Password: pass,
			}),
		}
		repoDst.PlainHTTP = true

		_, err = oras.Copy(
			ctx,
			repoSrc, repoSrc.Reference.Reference,
			repoDst, repoDst.Reference.Reference,
			oras.CopyOptions{},
		)
		r.NoError(err, "failed to copy CLI image to local registry")

		By("writing the component constructor definition")
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
`,
			componentName,
			componentVersion,
			resourceName,
			componentVersion,
			"http://"+dstRef,
		)

		constructorPath := filepath.Join(tempDir, "constructor.yaml")
		r.NoError(os.WriteFile(constructorPath, []byte(constructorContent), os.ModePerm))

		By("writing OCM credential configuration for the local registry")
		host, port, err := net.SplitHostPort(registryAddress)
		r.NoError(err)

		ocmConfigContent := fmt.Sprintf(`
type: generic.config.ocm.software/v1
configurations:
- type: credentials.config.ocm.software
  consumers:
    - identity:
        type: OCIRepository
        hostname: %q
        port: %q
        scheme: http
      credentials:
      - type: Credentials/v1
        properties:
          username: %s
          password: %s
`, host, port, user, pass)

		configPath := filepath.Join(tempDir, "ocmconfig.yaml")
		r.NoError(os.WriteFile(configPath, []byte(strings.TrimSpace(ocmConfigContent)), os.ModePerm))

		By("adding the component version to the registry using the OCM CLI")
		addArgs := []string{
			"run", "--rm",
			"--network", "host",
			"-v", fmt.Sprintf("%s:/work", tempDir),
			"ghcr.io/open-component-model/cli:main",
			"add", "component-version",
			"--repository", "http://" + registryAddress,
			"--config", "/work/ocmconfig.yaml",
			"--constructor", "/work/constructor.yaml",
			"--loglevel", "debug",
		}

		cmd := exec.Command("docker", addArgs...)
		out, err := cmd.CombinedOutput()
		r.NoError(err, string(out))

		By("verifying the component version exists in the registry")
		getArgs := []string{
			"run", "--rm",
			"--network", "host",
			"-v", fmt.Sprintf("%s:/work", tempDir),
			"ghcr.io/open-component-model/cli:main",
			"get", "component-version",
			"--config", "/work/ocmconfig.yaml",
			"-oyaml",
			fmt.Sprintf("http://%s//%s:%s", registryAddress, componentName, componentVersion),
		}

		cmd = exec.Command("docker", getArgs...)
		out, err = cmd.CombinedOutput()
		r.NoError(err, string(out))

		credArgs := []string{
			"--cred", ":type=OCIRegistry",
			"--cred", ":hostname=" + host,
			"--cred", ":port=" + port,
			"--cred", ":scheme=http",
			"--cred", "username=" + user,
			"--cred", "password=" + pass,
		}

		By("ensuring the resource is accessible via the OCM CLI")
		Expect(env.Execute(append(credArgs,
			"get", "resource",
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress,
			"-ojson",
			componentName+":"+componentVersion, resourceName,
		)...)).To(Succeed())

		By("downloading the resource to verify on-demand synthesis")
		Expect(env.Execute(append(credArgs,
			"download", "resource",
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress,
			componentName+":"+componentVersion, resourceName,
			"--outfile", filepath.Join(tempDir, "downloaded-cli"),
		)...)).To(Succeed())

		By("transferring the component version to a CTF with localBlob synthesis")
		targetCTF := filepath.Join(tempDir, "target-ctf")
		Expect(env.Execute(append(credArgs,
			"transfer", "componentversions",
			componentName+":"+componentVersion,
			targetCTF,
			"--repo", "http://"+registryAddress,
			"--lookup", "http://"+registryAddress,
			"--copy-resources",
		)...)).To(Succeed())

		By("verifying the resource is stored as a localBlob in the CTF")
		repo, err := ctf.Open(env, accessobj.ACC_READONLY, targetCTF, 0o700, ctf.FormatDirectory)
		r.NoError(err)
		defer repo.Close()

		cv, err := repo.LookupComponentVersion(componentName, componentVersion)
		r.NoError(err)
		defer cv.Close()

		var res cpi.ResourceAccess
		for _, rsc := range cv.GetResources() {
			if rsc.Meta().Name == resourceName {
				res = rsc
				break
			}
		}
		r.NotNil(res)

		meth, err := res.AccessMethod()
		r.NoError(err)
		defer meth.Close()

		r.Equal("localBlob", meth.GetKind())

		By("extracting and resolving the OCI layout from the localBlob")
		reader, err := meth.Reader()
		r.NoError(err)
		defer reader.Close()

		tempfs, err := tarutils.ExtractTgzToTempFs(reader)
		r.NoError(err)

		store, err := oci.NewFromFS(ctx, vfs.AsIoFS(tempfs))
		r.NoError(err)

		desc, err = store.Resolve(ctx, "latest")
		r.NoError(err)
		r.NotNil(desc)
	})
})
