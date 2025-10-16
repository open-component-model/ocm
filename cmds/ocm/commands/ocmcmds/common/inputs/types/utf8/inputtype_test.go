package utf8

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	. "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/testutils"
)

var _ = Describe("Input Type", func() {
	Context("options", func() {
		var env *InputTest

		BeforeEach(func() {
			env = NewInputTest(TYPE)
		})

		It("simple string decode", func() {
			env.Set(options.CompressOption, "true")
			env.Set(options.MediaTypeOption, "media")
			env.Set(options.TextOption, "stringdata")
			env.Check(&Spec{
				Text:        "stringdata",
				ProcessSpec: cpi.NewProcessSpec("media", true),
			})
		})

		It("simple json decode", func() {
			env.Set(options.CompressOption, "true")
			env.Set(options.MediaTypeOption, "media")
			env.Set(options.JSONOption, `field: value`)
			env.Check(&Spec{
				Json:        []byte(`{"field":"value"}`),
				ProcessSpec: cpi.NewProcessSpec("media", true),
			})
		})

		It("simple formatted json decode", func() {
			env.Set(options.CompressOption, "true")
			env.Set(options.MediaTypeOption, "media")
			env.Set(options.FormattedJSONOption, `field: value`)
			env.Check(&Spec{
				FormattedJson: []byte(`{"field":"value"}`),
				ProcessSpec:   cpi.NewProcessSpec("media", true),
			})
		})

		It("simple yaml decode", func() {
			env.Set(options.CompressOption, "true")
			env.Set(options.MediaTypeOption, "media")
			env.Set(options.YAMLOption, `field: value`)
			env.Check(&Spec{
				Yaml:        []byte(`{"field":"value"}`),
				ProcessSpec: cpi.NewProcessSpec("media", true),
			})
		})
	})

	Context("blob", func() {
		ctx := inputs.NewContext(clictx.DefaultContext(), nil, nil)

		It("handles text", func() {
			inp := New("stringdata", "media", false)

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte("stringdata")))
			Expect(info).To(Equal(""))
		})

		It("handles json from string", func() {
			inp := Must(NewJson("field: value", "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`{"field":"value"}`)))
			Expect(info).To(Equal(""))
		})
		It("handles json", func() {
			inp := Must(NewJson(map[string]interface{}{"field": "value"}, "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`{"field":"value"}`)))
			Expect(info).To(Equal(""))
		})

		It("handles formatted json from string", func() {
			inp := Must(NewFormattedJson("field: value", "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`{
  "field": "value"
}`)))
			Expect(info).To(Equal(""))
		})
		It("handles formatted json", func() {
			inp := Must(NewFormattedJson(map[string]interface{}{"field": "value"}, "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`{
  "field": "value"
}`)))
			Expect(info).To(Equal(""))
		})

		It("handles yaml from string", func() {
			inp := Must(NewYaml("field: value", "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`field: value
`)))
			Expect(info).To(Equal(""))
		})
		It("handles yaml", func() {
			inp := Must(NewYaml(map[string]interface{}{"field": "value"}, "media", false))

			a, info := Must2(inp.GetBlob(ctx, inputs.InputResourceInfo{}))
			Expect(a.MimeType()).To(Equal("media"))
			Expect(a.Get()).To(Equal([]byte(`field: value
`)))
			Expect(info).To(Equal(""))
		})
	})
})
