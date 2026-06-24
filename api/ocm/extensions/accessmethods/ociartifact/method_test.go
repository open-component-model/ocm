package ociartifact_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
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

	Context("tag only", func() {
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
			Expect(accspeccpi.GetAccessMethodImplementation(m).(blobaccess.DigestSource).Digest().String()).To(Equal("sha256:" + D_OCIMANIFEST1))
		})

		It("provides artifact hint", func() {
			spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION))

			hint := spec.GetReferenceHint(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(hint).To(Equal("ocm/value:v2.0"))
		})
	})

	Context("tag + digest", func() {
		It("accesses artifact", func() {
			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env)
			})

			FakeOCIRepo(env, OCIPATH, OCIHOST)

			spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION+"@sha256:"+D_OCIMANIFEST1))

			m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(err).To(Succeed())

			// no credentials required for CTF as fake OCI registry.
			Expect(credentials.GetProvidedConsumerId(m)).To(BeNil())
			Expect(accspeccpi.GetAccessMethodImplementation(m).(blobaccess.DigestSource).Digest().String()).To(Equal("sha256:" + D_OCIMANIFEST1))
		})

		It("provides artifact hint", func() {
			spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION+"@sha256:"+D_OCIMANIFEST1))

			hint := spec.GetReferenceHint(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(hint).To(Equal("ocm/value:v2.0"))
		})
	})

	Context("digest", func() {
		It("accesses artifact", func() {
			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env)
			})

			FakeOCIRepo(env, OCIPATH, OCIHOST)

			spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1))

			m, err := spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(err).To(Succeed())

			// no credentials required for CTF as fake OCI registry.
			Expect(credentials.GetProvidedConsumerId(m)).To(BeNil())
			Expect(accspeccpi.GetAccessMethodImplementation(m).(blobaccess.DigestSource).Digest().String()).To(Equal("sha256:" + D_OCIMANIFEST1))
		})

		It("provides artifact hint", func() {
			spec := ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1))

			hint := spec.GetReferenceHint(&cpi.DummyComponentVersionAccess{env.OCMContext()})
			Expect(hint).To(Equal("ocm/value"))
		})
	})
})

// Alias kinds cover the full set of names the OCM spec lists for this access
// type (doc/04-extensions/02-access-types/ociartifact.md): `OCIImage` is the
// canonical name, `ociArtifact` / `ociRegistry` / `ociImage` are legacy
// aliases. Each must decode (without erroring), be recognised by
// `ociartifact.Is`, preserve the payload, and marshal without panicking.
// Unlike `localblob`, the global decoder normalises the `type` token to a
// single registered kind, so we do not assert exact token preservation here.
var _ = Describe("Alias kinds", func() {
	const imageRef = "ghcr.io/example/image:v1"

	mkData := func(typ string) []byte {
		return []byte(fmt.Sprintf(`{"type":%q,"imageReference":%q}`, typ, imageRef))
	}

	DescribeTable("decode, marshal and Is() for every kind variant",
		func(typ string) {
			in := mkData(typ)

			By("decoding via the global ocm context")
			spec := Must(ocm.DefaultContext().AccessSpecForConfig(in, nil))
			Expect(spec).To(BeAssignableToTypeOf(&ociartifact.AccessSpec{}))
			Expect(spec.(*ociartifact.AccessSpec).ImageReference).To(Equal(imageRef))

			By("recognising the spec through ociartifact.Is")
			Expect(ociartifact.Is(spec)).To(BeTrue())

			By("marshalling without panicking")
			out := Must(json.Marshal(spec))
			Expect(out).NotTo(BeEmpty())

			By("round-tripping the marshalled form back to an AccessSpec")
			again := Must(ocm.DefaultContext().AccessSpecForConfig(out, nil))
			Expect(again).To(BeAssignableToTypeOf(&ociartifact.AccessSpec{}))
			Expect(again.(*ociartifact.AccessSpec).ImageReference).To(Equal(imageRef))
			Expect(ociartifact.Is(again)).To(BeTrue())
		},
		Entry("OCIImage canonical", ociartifact.LegacyType2),
		Entry("OCIImage canonical /v1", ociartifact.LegacyType2V1),
		Entry("ociArtifact legacy alias", ociartifact.Type),
		Entry("ociArtifact legacy alias /v1", ociartifact.TypeV1),
		Entry("ociRegistry legacy alias", ociartifact.LegacyType),
		Entry("ociRegistry legacy alias /v1", ociartifact.LegacyTypeV1),
		Entry("ociImage legacy alias", ociartifact.LegacyType3),
		Entry("ociImage legacy alias /v1", ociartifact.LegacyType3V1),
	)
})
