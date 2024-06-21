package main_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
)

var _ = Describe("OCM command line test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	// https://github.com/open-component-model/ocm-project/issues/22
	It("Add OCI image resource fails when a cache directory is specified on Windows", func() {
		// Set ocmconfig.CacheDir
		Expect(env.Execute("config", "set", "cacheDir", "/Temp")).To(Succeed())

		// ocm create ca --file ca --scheme ocm.software/v3alpha1 --provider test.com test.com/postgresql 14.0.5
		Expect(env.Execute("create", "ca", "--file", "ca", "--scheme", "ocm.software/v3alpha1", "--provider", "test.com", "test.com/postgresql", "14.0.5")).To(Succeed())

		// ocm add resource --file ca --name bitnami-shell --type ociImage --accessType ociArtifact --version 10 --reference bitnami/postgresql:16.2.0-debian-11-r1
		Expect(env.Execute("add", "resource", "--file", "ca", "--name", "bitnami-shell", "--type", "ociImage", "--accessType", "ociArtifact", "--version", "10", "--reference", "bitnami/postgresql:16.2.0-debian-11-r1")).To(Succeed())
	})
})
