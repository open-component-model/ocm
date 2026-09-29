package wget_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/ocm/extensions/accessmethods/wget"

	"ocm.software/ocm/api/ocm"
	_ "ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
)

const aliasURL = "https://example.com/file"

var _ = Describe("wget access type spellings", func() {
	mkData := func(typ string) []byte {
		return []byte(fmt.Sprintf(`{"type":%q,"URL":%q}`, typ, aliasURL))
	}

	DescribeTable("decode and round-trip every accepted type spelling",
		func(typ string) {
			in := mkData(typ)

			By("decoding via the global ocm context")
			spec := Must(ocm.DefaultContext().AccessSpecForConfig(in, nil))
			Expect(spec).To(BeAssignableToTypeOf(&AccessSpec{}))
			Expect(spec.GetType()).To(Equal(typ))
			Expect(spec.(*AccessSpec).URL).To(Equal(aliasURL))

			By("recognising the spec through Is")
			Expect(Is(spec)).To(BeTrue())

			By("marshalling preserving the type token")
			out := Must(json.Marshal(spec))
			Expect(string(out)).To(ContainSubstring(fmt.Sprintf(`"type":%q`, typ)))

			By("decoding the marshalled form again")
			redecoded := Must(ocm.DefaultContext().AccessSpecForConfig(out, nil))
			Expect(redecoded).To(BeAssignableToTypeOf(&AccessSpec{}))
			Expect(redecoded.GetType()).To(Equal(typ))
			Expect(redecoded.(*AccessSpec).URL).To(Equal(aliasURL))
		},
		Entry("wget", "wget"),
		Entry("wget/v1", "wget/v1"),
		Entry("Wget", "Wget"),
		Entry("Wget/v1", "Wget/v1"),
		Entry("http", "http"),
		Entry("http/v1", "http/v1"),
		Entry("HTTP", "HTTP"),
		Entry("HTTP/v1", "HTTP/v1"),
	)

	It("returns false for nil", func() {
		Expect(Is(nil)).To(BeFalse())
	})

	It("does not falsely claim other access types", func() {
		foreign := Must(ocm.DefaultContext().AccessSpecForConfig([]byte(
			`{"type":"ociArtifact/v1","imageReference":"ghcr.io/foo/bar:1.0.0"}`), nil))
		Expect(Is(foreign)).To(BeFalse())
	})
})
