package plugin_test

import (
	"os"

	"github.com/mandelsoft/goutils/sliceutils"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/goutils/transformer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/options"
	. "ocm.software/ocm/api/ocm/plugin/testutils"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	. "ocm.software/ocm/cmds/ocm/testhelper"

	"github.com/mandelsoft/filepath/pkg/filepath"

	"ocm.software/ocm/api/ocm/extensions/attrs/plugincacheattr"
	"ocm.software/ocm/api/ocm/plugin/plugins"
	"ocm.software/ocm/api/utils/runtime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/plugin"
)

const PLUGIN = "input"

const TESTDATA = "this is some test data\n"

var _ = Describe("Input Command Test Environment", func() {
	Context("plugin execution", func() {
		var env *TestEnv
		var plugindir TempPluginDir
		var registry plugins.Set

		BeforeEach(func() {
			env = NewTestEnv(TestData())
			plugindir = Must(ConfigureTestPlugins(env, "testdata/plugins"))
			registry = plugincacheattr.Get(env)
		})

		AfterEach(func() {
			plugindir.Cleanup()
			env.Cleanup()
		})

		It("loads plugin", func() {
			//	Expect(registration.RegisterExtensions(env)).To(Succeed())
			p := registry.Get(PLUGIN)
			Expect(p).NotTo(BeNil())
			Expect(p.Error()).To(Equal(""))
		})

		It("gets blob", func() {
			p := registry.Get(PLUGIN)
			t := plugin.NewType("demo", p, &p.GetDescriptor().Inputs[0])

			file := Must(os.CreateTemp("", "input*"))
			defer os.Remove(file.Name())
			Must(file.Write([]byte(TESTDATA)))
			file.Close()
			spec := `
type: demo
path: ` + filepath.Base(file.Name()) + `
mediaType: plain/text
`
			is := Must(t.Decode([]byte(spec), runtime.DefaultYAMLEncoding))
			Expect(is.GetType()).To(Equal("demo"))

			ctx := inputs.NewContext(env.CLIContext(), nil, nil)
			blob, hint := Must2(is.GetBlob(ctx, inputs.InputResourceInfo{}))
			defer blob.Close()
			Expect(hint).To(Equal(filepath.Base(file.Name())))
			data := Must(blob.Get())
			Expect(string(data)).To(Equal(TESTDATA))
		})

		It("gets input", func() {
			scheme := inputs.For(env.CLIContext())
			plugin.RegisterPlugins(env.CLIContext())

			file := Must(os.CreateTemp("", "input*"))
			defer os.Remove(file.Name())
			Must(file.Write([]byte(TESTDATA)))
			file.Close()
			spec := `
type: demo
path: ` + filepath.Base(file.Name()) + `
mediaType: plain/text
`

			is := Must(scheme.DecodeInputSpec([]byte(spec), runtime.DefaultYAMLEncoding))
			Expect(is.GetType()).To(Equal("demo"))

			ctx := inputs.NewContext(env.CLIContext(), nil, nil)
			blob, hint := Must2(is.GetBlob(ctx, inputs.InputResourceInfo{}))
			defer blob.Close()
			Expect(hint).To(Equal(filepath.Base(file.Name())))
			data := Must(blob.Get())
			Expect(string(data)).To(Equal(TESTDATA))
		})

		It("handles input options", func() {
			scheme := inputs.For(env.CLIContext())
			plugin.RegisterPlugins(env.CLIContext())

			it := scheme.GetInputType("demo")
			Expect(it).NotTo(BeNil())

			h := it.ConfigOptionTypeSetHandler()
			Expect(h).NotTo(BeNil())
			Expect(h.GetName()).To(Equal("demo"))

			ot := h.OptionTypes()
			Expect(len(ot)).To(Equal(2))

			opts := h.CreateOptions()
			Expect(sliceutils.Transform(opts.Options(), transformer.GetName[flagsets.Option, string])).To(ConsistOf(
				"mediaType", "inputPath"))

			fs := &pflag.FlagSet{}
			fs.SortFlags = true
			opts.AddFlags(fs)

			Expect("\n" + fs.FlagUsages()).To(Equal(`
      --inputPath string   path field for input
      --mediaType string   media type for artifact blob representation
`))

			MustBeSuccessful(fs.Parse([]string{"--inputPath", "filepath", "--" + options.MediatypeOption.GetName(), "yaml"}))

			cfg := flagsets.Config{}
			MustBeSuccessful(h.ApplyConfig(opts, cfg))
			Expect(cfg).To(YAMLEqual(`
mediaType: yaml
path: filepath
`))
		})
	})
})
