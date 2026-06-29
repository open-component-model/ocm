package localblob_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"

	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/api/utils/runtime"
)

const (
	CTF              = "ctf"
	COMPONENT        = "fabianburth.org/component"
	VERSION          = "v1.0"
	ARTIFACT_NAME    = "artifact"
	ARTIFACT_VERSION = "v1.0"
)

var _ = Describe("Method", func() {
	data := `globalAccess:
  digest: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
  mediaType: application/tar+gzip
  ref: ghcr.io/vasu1124/ocm/component-descriptors/github.com/vasu1124/introspect-delivery
  size: 11287
  type: ociBlob
localReference: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
mediaType: application/tar+gzip
type: localBlob
`
	_ = data

	It("marshal/unmarshal simple", func() {
		spec := localblob.New("path", "hint", mime.MIME_TEXT, nil)
		data := Must(json.Marshal(spec))
		Expect(string(data)).To(Equal("{\"type\":\"localBlob\",\"localReference\":\"path\",\"mediaType\":\"text/plain\",\"referenceName\":\"hint\"}"))
		r := Must(localblob.Decode(data))
		Expect(r).To(Equal(spec))
	})

	It("marshal/unmarshal with global", func() {
		spec := localblob.New("", "", "", nil)
		Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), spec)).To(Succeed())

		r := Must(runtime.DefaultYAMLEncoding.Marshal(spec))
		Expect(string(r)).To(Equal(data))

		global := ociblob.New(
			"ghcr.io/vasu1124/ocm/component-descriptors/github.com/vasu1124/introspect-delivery",
			"sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a",
			"application/tar+gzip",
			11287,
		)
		Expect(spec.GlobalAccess.Evaluate(ocm.DefaultContext())).To(Equal(global))

		r = Must(runtime.DefaultYAMLEncoding.Marshal(spec))
		Expect(string(r)).To(Equal(data))
	})

	It("check get inexpensive content version identity method", func() {
		var env *Builder

		env = NewBuilder()
		defer env.Cleanup()

		env.OCMCommonTransport(CTF, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENT, VERSION, func() {
				env.Resource(ARTIFACT_NAME, ARTIFACT_VERSION, resourcetypes.BLOB, metav1.LocalRelation, func() {
					env.BlobData(mime.MIME_TEXT, []byte("testdata"))
				})
			})
		})

		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, CTF, 0, env))
		defer Close(repo)
		cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv)
		access := cv.GetDescriptor().Resources[0].Access
		spec := Must(env.OCMContext().AccessSpecForSpec(access))
		Expect(spec.GetVersion()).To(Equal("v1"))
	})
})

// Alias kinds cover the v2-emitted variants registered through the local versions
// scheme (see method.go, issue #1979). Each variant must decode through the global
// context and the package-level Decode, marshal without panicking, and round-trip
// while preserving the literal type token. Pre-fix, the upper-cased variants
// panicked at MarshalJSON because their encoder was the global scheme, which has
// no converter registered for them.
var _ = Describe("Alias kinds", func() {
	mkData := func(typ string) []byte {
		return []byte(fmt.Sprintf(
			`{"type":%q,"localReference":"path","mediaType":"text/plain","referenceName":"hint"}`,
			typ,
		))
	}

	DescribeTable("decode and round-trip every kind variant",
		func(typ string) {
			in := mkData(typ)

			By("decoding via the global ocm context")
			spec := Must(ocm.DefaultContext().AccessSpecForConfig(in, nil))
			Expect(spec).To(BeAssignableToTypeOf(&localblob.AccessSpec{}))
			Expect(spec.GetType()).To(Equal(typ))

			By("recognising the spec through localblob.Is")
			Expect(localblob.Is(spec)).To(BeTrue())

			By("decoding via the package-level Decode (private versions scheme)")
			pkgSpec := Must(localblob.Decode(in))
			Expect(pkgSpec).To(BeAssignableToTypeOf(&localblob.AccessSpec{}))
			Expect(pkgSpec.GetType()).To(Equal(typ))

			By("marshalling without panicking and preserving the type token")
			out := Must(json.Marshal(spec))
			Expect(string(out)).To(ContainSubstring(fmt.Sprintf(`"type":%q`, typ)))

			By("round-tripping the marshalled form back to the same type")
			again := Must(ocm.DefaultContext().AccessSpecForConfig(out, nil))
			Expect(again).To(BeAssignableToTypeOf(&localblob.AccessSpec{}))
			Expect(again.GetType()).To(Equal(typ))
		},
		Entry("canonical", localblob.Type),
		Entry("canonical /v1", localblob.TypeV1),
		Entry("upper-case (v2 alias)", localblob.UpperType),
		Entry("upper-case /v1 (v2 alias)", localblob.UpperTypeV1),
	)
})
