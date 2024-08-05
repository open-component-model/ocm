package ociartifact_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/accspeccpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
)

const (
	OCIPATH = "/tmp/oci"
	OCIHOST = "alias"
)

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("accesses artifact", func() {
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION))

		m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
		Expect(err).To(Succeed())

		// no credentials required for CTF as fake OCI registry.
		Expect(credentials.GetProvidedConsumerId(m)).To(BeNil())
		Expect(accspeccpi.GetAccessMethodImplementation(m).(blobaccess.DigestSource).Digest().String()).To(Equal("sha256:0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f"))
	})
})
