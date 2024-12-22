package common_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common"

	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/addhdlrs"
)

var _ = Describe("Blob Inputs", func() {
	It("missing input", func() {
		in := `
access:
  type: localBlob
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("simple decode", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})
	It("complains about additional input field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  bla: blub
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err.Error()).To(Equal("input.bla: Forbidden: unknown field"))
	})

	It("does not complains about additional dir field", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: dir
  excludeFiles:
     - xyz
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(Succeed())
	})

	It("complains about additional dir field for file", func() {
		in := `
access:
  type: localBlob
input:
  mediaType: text/plain
  path: test
  type: file
  excludeFiles:
  - xyz
`
		_, err := addhdlrs.DecodeInput([]byte(in), nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("input.excludeFiles: Forbidden: unknown field"))
	})

	FContext("resource provider", func() {
		var prov *common.ContentResourceSpecificationsProvider
		var fs *pflag.FlagSet

		BeforeEach(func() {
			prov = common.NewContentResourceSpecificationProvider(clictx.DefaultContext(), "test", nil, "deftype")
			fs = &pflag.FlagSet{}
			prov.AddFlags(fs)
		})

		It("handles single hint", func() {
			MustBeSuccessful(fs.Parse([]string{
				"--refhint",
				"type=oci",
				"--refhint",
				"reference=ref",
			}))

			meta := Must(prov.ParsedMeta())
			Expect(meta).To(YAMLEqual(`
 type: deftype
 referenceHints:
  - reference: ref
    type: oci
`))
		})

		It("handles multiple hints", func() {
			MustBeSuccessful(fs.Parse([]string{
				"--refhint",
				"type=oci",
				"--refhint",
				"reference=ref",
				"--refhint",
				"type=helm",
				"--refhint",
				"reference=chart",
			}))

			meta := Must(prov.ParsedMeta())
			Expect(meta).To(YAMLEqual(`
 type: deftype
 referenceHints:
  - reference: ref
    type: oci
  - reference: chart
    type: helm
`))
		})

		It("handles multiple simple hints", func() {
			MustBeSuccessful(fs.Parse([]string{
				"--refhint",
				"oci::ref",
				"--refhint",
				"helm::chart",
			}))

			meta := Must(prov.ParsedMeta())
			Expect(meta).To(YAMLEqual(`
 type: deftype
 referenceHints:
  - reference: ref
    type: oci
  - reference: chart
    type: helm
`))

		})
	})
})
