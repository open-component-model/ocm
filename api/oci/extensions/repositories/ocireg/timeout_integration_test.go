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

		registryContainer, err = registry.Run(ctx, registryImage,
			network.WithNetwork([]string{"registry"}, nw),
			testcontainers.WithWaitStrategy(wait.ForHTTP("/v2/").WithPort("5000/tcp")),
		)
		Expect(err).To(Succeed(), "failed to start registry container")

		toxiContainer, err = tctoxiproxy.Run(ctx, toxiproxyImage,
			network.WithNetwork([]string{"toxiproxy"}, nw),
			tctoxiproxy.WithProxy(proxyName, fmt.Sprintf("registry:%d", registryPort)),
		)
		Expect(err).To(Succeed(), "failed to start toxiproxy container")

		host, port, err := toxiContainer.ProxiedEndpoint(8666)
		Expect(err).To(Succeed())
		proxyHost = fmt.Sprintf("%s:%s", host, port)

		uri, err := toxiContainer.URI(ctx)
		Expect(err).To(Succeed())
		toxiClient := toxiproxy.NewClient(uri)
		proxy, err = toxiClient.Proxy(proxyName)
		Expect(err).To(Succeed(), "failed to get toxiproxy proxy")

		env = NewTestEnv(envhelper.FileSystem(osfs.New()))

		// Create temp dir and CTF for all tests.
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

	// writeHTTPConfig writes a config file with the given http settings and returns its path.
	writeHTTPConfig := func(settings string) string {
		cfg := fmt.Sprintf(`{"type":"http.config.ocm.software/v1alpha1",%s}`, settings)
		cfgFile := filepath.Join(tempDir, "httpconfig.yaml")
		Expect(os.WriteFile(cfgFile, []byte(cfg), 0o644)).To(Succeed())
		return cfgFile
	}

	It("fails when timeout is shorter than proxy latency", func() {
		addLatency(proxy, 30_000, "upstream")
		defer removeToxic(proxy, "latency")

		cfgFile := writeHTTPConfig(`"timeout":"2s"`)
		registryURL := "http://" + proxyHost
		err := env.Execute(
			"--config", cfgFile,
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
		addLatency(proxy, 1_000, "upstream")
		defer removeToxic(proxy, "latency")

		cfgFile := writeHTTPConfig(`"timeout":"30s"`)
		registryURL := "http://" + proxyHost
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"--config", cfgFile,
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)).To(Succeed())
	})

	It("fails when response headers can't arrive within configured time", func() {
		addLatency(proxy, 5_000, "downstream")
		defer removeToxic(proxy, "latency")

		// Cap overall timeout to prevent retries from dragging out.
		cfgFile := writeHTTPConfig(`"responseHeaderTimeout":"1s","timeout":"8s"`)
		registryURL := "http://" + proxyHost
		err := env.Execute(
			"--config", cfgFile,
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(SatisfyAny(
			ContainSubstring("timeout awaiting response headers"),
			ContainSubstring("context deadline exceeded"),
			ContainSubstring("i/o timeout"),
		))
	})

	It("succeeds when response header timeout is generous enough", func() {
		addLatency(proxy, 2_000, "downstream")
		defer removeToxic(proxy, "latency")

		cfgFile := writeHTTPConfig(`"responseHeaderTimeout":"30s"`)
		registryURL := "http://" + proxyHost
		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute(
			"--config", cfgFile,
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)).To(Succeed())
	})

	It("fails when tcp-dial-timeout is too short to establish connection", func() {
		cfgFile := writeHTTPConfig(`"tcpDialTimeout":"1ns"`)
		registryURL := "http://" + proxyHost
		err := env.Execute(
			"--config", cfgFile,
			"transfer", "componentversions",
			ctfDir+"//"+componentName+":"+componentVersion,
			registryURL,
		)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(SatisfyAny(
			ContainSubstring("i/o timeout"),
			ContainSubstring("dial tcp"),
			ContainSubstring("context deadline exceeded"),
		))
	})

})

func addLatency(proxy *toxiproxy.Proxy, latencyMs int, stream string) {
	_, err := proxy.AddToxic("latency", "latency", stream, 1.0, toxiproxy.Attributes{
		"latency": latencyMs,
	})
	Expect(err).To(Succeed())
}

func removeToxic(proxy *toxiproxy.Proxy, name string) {
	Expect(proxy.RemoveToxic(name)).To(Succeed())
}
