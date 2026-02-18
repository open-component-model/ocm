//go:build integration

package ocireg_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	toxiproxy "github.com/Shopify/toxiproxy/v2/client"
	"github.com/mandelsoft/vfs/pkg/osfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/registry"
	tctoxiproxy "github.com/testcontainers/testcontainers-go/modules/toxiproxy"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	envhelper "ocm.software/ocm/api/helper/env"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

var _ = Describe("Registry timeout:", Ordered, func() {
	const (
		registryImage    = "registry:2.8.3"
		toxiproxyImage   = "ghcr.io/shopify/toxiproxy:2.12.0"
		componentName    = "example.com/timeout-test"
		componentVersion = "1.0.0"
		proxyName        = "registry"
		registryPort     = 5000
	)

	var (
		ctx               context.Context
		nw                *testcontainers.DockerNetwork
		registryContainer testcontainers.Container
		toxiContainer     *tctoxiproxy.Container
		proxy             *toxiproxy.Proxy
		proxyHost         string
		tempDir           string
		ctfDir            string
		env               *TestEnv
	)

	BeforeAll(func() {
		ctx = context.Background()
		log := GinkgoLogr

		var err error
		nw, err = network.New(ctx)
		Expect(err).To(Succeed(), "failed to create Docker network")

		// Start registry on the shared network with alias "registry".
		registryContainer, err = registry.Run(ctx, registryImage,
			network.WithNetwork([]string{"registry"}, nw),
			testcontainers.WithWaitStrategy(wait.ForHTTP("/v2/").WithPort("5000/tcp")),
		)
		Expect(err).To(Succeed(), "failed to start registry container")

		// Start toxiproxy on the same network with a proxy forwarding to the registry.
		toxiContainer, err = tctoxiproxy.Run(ctx, toxiproxyImage,
			network.WithNetwork([]string{"toxiproxy"}, nw),
			tctoxiproxy.WithProxy(proxyName, fmt.Sprintf("registry:%d", registryPort)),
		)
		Expect(err).To(Succeed(), "failed to start toxiproxy container")

		// Get the host-mapped endpoint for the proxy.
		host, port, err := toxiContainer.ProxiedEndpoint(8666)
		Expect(err).To(Succeed())
		proxyHost = fmt.Sprintf("%s:%s", host, port)

		// Obtain a reference to the proxy for adding toxics.
		uri, err := toxiContainer.URI(ctx)
		Expect(err).To(Succeed())
		toxiClient := toxiproxy.NewClient(uri)
		proxy, err = toxiClient.Proxy(proxyName)
		Expect(err).To(Succeed(), "failed to get toxiproxy proxy")

		env = NewTestEnv(envhelper.FileSystem(osfs.New()))

		log.Info("Toxic Registry ready", "proxy", proxyHost, "ctf", ctfDir)
	})

	AfterAll(func() {
		if toxiContainer != nil {
			Expect(testcontainers.TerminateContainer(toxiContainer)).To(Succeed())
		}
		if registryContainer != nil {
			Expect(testcontainers.TerminateContainer(registryContainer)).To(Succeed())
		}
		if nw != nil {
			Expect(nw.Remove(ctx)).To(Succeed())
		}
		if env != nil {
			Expect(env.Cleanup()).To(Succeed())
		}
		if tempDir != "" {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

	It("creates component version", func() {
		// Build a minimal component version (CTF) on disk.
		var err error
		tempDir, err = os.MkdirTemp("", "ocm-timeout-*")
		Expect(err).To(Succeed())

		ctfDir = filepath.Join(tempDir, "ctf")
		constructorFile := filepath.Join(tempDir, "constructor.yaml")
		constructor := `components:
  - name: ` + componentName + `
    version: ` + componentVersion + `
    provider:
      name: test
`
		Expect(os.WriteFile(constructorFile, []byte(constructor), 0o644)).To(Succeed())

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"add", "componentversions",
			"--create",
			"--file", ctfDir,
			constructorFile,
		)).To(Succeed())

	})

	It("fails when timeout is shorter than proxy latency", func() {
		// Add 30s latency
		_, err := proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
			"latency": 30_000,
		})
		Expect(err).To(Succeed())
		defer func() {
			Expect(proxy.RemoveToxic("latency")).To(Succeed())
		}()

		registryURL := "http://" + proxyHost
		err = env.Execute(
			"--timeout", "2s",
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(SatisfyAny(
			ContainSubstring("Client.Timeout"),
			ContainSubstring("context deadline exceeded"),
			ContainSubstring("i/o timeout"),
		))
	})

	It("fails with default timeout when latency exceeds 30s", func() {
		// Add 60s latency — exceeds the default 30s timeout.
		_, err := proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
			"latency": 60_000,
		})
		Expect(err).To(Succeed())
		defer func() {
			Expect(proxy.RemoveToxic("latency")).To(Succeed())
		}()

		registryURL := "http://" + proxyHost
		err = env.Execute(
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(SatisfyAny(
			ContainSubstring("Client.Timeout"),
			ContainSubstring("context deadline exceeded"),
			ContainSubstring("i/o timeout"),
		))
	})

	It("succeeds when timeout exceeds proxy latency", func() {
		// Add 1s latency
		_, err := proxy.AddToxic("latency", "latency", "upstream", 1.0, toxiproxy.Attributes{
			"latency": 1_000,
		})
		Expect(err).To(Succeed())
		defer func() {
			Expect(proxy.RemoveToxic("latency")).To(Succeed())
		}()
		registryURL := "http://" + proxyHost
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"--timeout", "30s",
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)).To(Succeed())
	})
})
