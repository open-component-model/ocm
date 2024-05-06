package ociartifact_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	. "github.com/open-component-model/ocm/pkg/env/builder"
)

const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

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
