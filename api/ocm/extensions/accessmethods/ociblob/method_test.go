package ociblob_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	"ocm.software/ocm/api/utils/accessio"
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
		var desc *artdesc.Descriptor
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			desc = OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		spec := ociblob.New(OCIHOST+".alias"+grammar.RepositorySeparator+OCINAMESPACE, desc.Digest, "", -1)

		m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
		Expect(err).To(Succeed())

		blob, err := m.Get()
		Expect(err).To(Succeed())

		Expect(string(blob)).To(Equal("manifestlayer"))
	})
})
