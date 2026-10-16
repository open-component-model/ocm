package wget_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/wget"
)

const aliasURL = "https://example.com/file"

var _ = Describe("wget input type spellings", func() {
	mkData := func(typ string) []byte {
		return []byte(fmt.Sprintf(`{"type":%q,"url":%q}`, typ, aliasURL))
	}

	DescribeTable("decode and round-trip every accepted type spelling",
		func(typ string) {
			in := mkData(typ)

			By("decoding via the default input type scheme")
			spec := Must(inputs.DefaultInputTypeScheme.DecodeInputSpec(in, runtime.DefaultJSONEncoding))
			Expect(spec).To(BeAssignableToTypeOf(&wget.Spec{}))
			Expect(spec.GetType()).To(Equal(typ))
			Expect(spec.(*wget.Spec).URL).To(Equal(aliasURL))

			By("marshalling preserving the type token")
			out := Must(json.Marshal(spec))
			Expect(string(out)).To(ContainSubstring(fmt.Sprintf(`"type":%q`, typ)))

			By("decoding the marshalled form again")
			redecoded := Must(inputs.DefaultInputTypeScheme.DecodeInputSpec(out, runtime.DefaultJSONEncoding))
			Expect(redecoded).To(BeAssignableToTypeOf(&wget.Spec{}))
			Expect(redecoded.GetType()).To(Equal(typ))
			Expect(redecoded.(*wget.Spec).URL).To(Equal(aliasURL))
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
})
