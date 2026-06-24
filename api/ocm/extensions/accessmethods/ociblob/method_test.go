package ociblob_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/grammar"
	"ocm.software/ocm/api/ocm"
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

// Alias kinds cover the names the OCM spec lists for this access type
// (doc/04-extensions/02-access-types/ociblob.md): `OCIImageLayer` is the
// canonical name, `ociBlob` is a legacy alias. Each must decode, be
// recognised as an *ociblob.AccessSpec, preserve its payload, and marshal
// without panicking. As with `ociartifact`, the global decoder normalises
// the `type` token to the registered kind, so we do not assert exact token
// preservation.
var _ = Describe("Alias kinds", func() {
	const (
		ref       = "ghcr.io/example/repo"
		mediaType = "application/octet-stream"
		dig       = "sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a"
	)

	mkData := func(typ string) []byte {
		return []byte(fmt.Sprintf(
			`{"type":%q,"ref":%q,"mediaType":%q,"digest":%q,"size":11287}`,
			typ, ref, mediaType, dig,
		))
	}

	DescribeTable("decode and marshal every kind variant",
		func(typ string) {
			in := mkData(typ)

			By("decoding via the global ocm context")
			spec := Must(ocm.DefaultContext().AccessSpecForConfig(in, nil))
			Expect(spec).To(BeAssignableToTypeOf(&ociblob.AccessSpec{}))

			as := spec.(*ociblob.AccessSpec)
			Expect(as.Reference).To(Equal(ref))
			Expect(as.MediaType).To(Equal(mediaType))
			Expect(as.Digest.String()).To(Equal(dig))
			Expect(as.Size).To(Equal(int64(11287)))

			By("marshalling without panicking")
			out := Must(json.Marshal(spec))
			Expect(out).NotTo(BeEmpty())

			By("round-tripping the marshalled form back to an AccessSpec")
			again := Must(ocm.DefaultContext().AccessSpecForConfig(out, nil))
			Expect(again).To(BeAssignableToTypeOf(&ociblob.AccessSpec{}))
			Expect(again.(*ociblob.AccessSpec).Reference).To(Equal(ref))
		},
		Entry("OCIImageLayer canonical", ociblob.UpperType),
		Entry("OCIImageLayer canonical /v1", ociblob.UpperTypeV1),
		Entry("ociBlob legacy alias", ociblob.Type),
		Entry("ociBlob legacy alias /v1", ociblob.TypeV1),
	)
})
