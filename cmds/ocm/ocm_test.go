package main_test

import (
	"os"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/oci/attrs/cacheattr"
)

var _ = Describe("OCM command line test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv(TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("Add OCI image resource fails when a cache directory is specified on Windows", func() {
		tmp := Must(os.MkdirTemp("", "ocm-cache-*"))
		defer os.RemoveAll(tmp)
		// configure ocm cache
		MustBeSuccessful(cacheattr.Set(env.OCIContext(), Must(accessio.NewStaticBlobCache(tmp))))

		// ocm create ca --file ca --scheme ocm.software/v3alpha1 --provider test.com test.com/postgresql 14.0.5
		Expect(env.Execute("create", "ca", "--file", "ca", "--scheme", "ocm.software/v3alpha1", "--provider", "test.com", "test.com/postgresql", "14.0.5")).To(Succeed())
		// ocm add resource --file ca --name bitnami-shell --type ociImage --accessType ociArtifact --version 10 --reference bitnami/postgresql:16.2.0-debian-11-r1
		Expect(env.Execute("add", "resource", "--file", "ca", "--name", "bitnami-shell", "--type", "ociImage", "--accessType", "ociArtifact", "--version", "10", "--reference", "bitnami/postgresql:16.2.0-debian-11-r1")).To(Succeed())
	})
})
