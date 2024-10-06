package inputs_test

import (
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/api/datacontext"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"ocm.software/ocm/api/utils/blobaccess/blobaccess"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/api/utils/mime"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/cpi"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
)

var _ = Describe("Input Type Extension Test Environment", func() {
	var (
		scheme inputs.InputTypeScheme
		itype  = inputs.NewInputType(TYPE, &Spec{}, "", ConfigHandler())
		flags  *pflag.FlagSet
		opts   flagsets.ConfigOptions
	)

	Context("registry", func() {
		BeforeEach(func() {
			scheme = inputs.NewInputTypeScheme(nil, inputs.DefaultInputTypeScheme)
			scheme.Register(itype)
			flags = &pflag.FlagSet{}
			opts = scheme.CreateConfigTypeSetConfigProvider().CreateOptions()
			opts.AddFlags(flags)
		})

		It("is not in base", func() {
			scheme = inputs.DefaultInputTypeScheme
			Expect(scheme.GetInputType(TYPE)).To(BeNil())
		})

		It("derives base input type", func() {
			prov := scheme.CreateConfigTypeSetConfigProvider()
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(prov.GetTypeOptionType(), ociartifact.TYPE),
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
				flagsets.OptionSpec(options.PlatformsOption, "linux/amd64"),
				flagsets.OptionSpec(options.PlatformsOption, "/arm64"),
			))
			cfg := Must(prov.GetConfigFor(opts))
			fmt.Printf("selected input options: %+v\n", cfg)

			spec := Must(scheme.GetInputSpecFor(cfg))
			Expect(spec).To(Equal(ociartifact.New("ghcr.io/open-component-model/image:v1.0", "linux/amd64", "/arm64")))
		})

		It("uses extended input type", func() {
			prov := scheme.CreateConfigTypeSetConfigProvider()
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(prov.GetTypeOptionType(), TYPE),
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
			))
			cfg := Must(prov.GetConfigFor(opts))
			fmt.Printf("selected input options: %+v\n", cfg)

			spec := Must(scheme.GetInputSpecFor(cfg))
			Expect(spec).To(Equal(New("ghcr.io/open-component-model/image:v1.0")))
		})
	})

	Context("cli context", func() {
		var ctx clictx.Context

		BeforeEach(func() {
			ctx = clictx.New(datacontext.MODE_EXTENDED)
			scheme = inputs.For(ctx)
			scheme.Register(itype)
			flags = &pflag.FlagSet{}
			opts = scheme.CreateConfigTypeSetConfigProvider().CreateOptions()
			opts.AddFlags(flags)
		})

		It("is not in base", func() {
			scheme = inputs.For(clictx.DefaultContext())
			Expect(scheme.GetInputType(TYPE)).To(BeNil())
		})

		It("derives base input type", func() {
			prov := scheme.CreateConfigTypeSetConfigProvider()
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(prov.GetTypeOptionType(), ociartifact.TYPE),
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
				flagsets.OptionSpec(options.PlatformsOption, "linux/amd64"),
				flagsets.OptionSpec(options.PlatformsOption, "/arm64"),
			))
			cfg := Must(prov.GetConfigFor(opts))
			fmt.Printf("selected input options: %+v\n", cfg)

			spec := Must(scheme.GetInputSpecFor(cfg))
			Expect(spec).To(Equal(ociartifact.New("ghcr.io/open-component-model/image:v1.0", "linux/amd64", "/arm64")))
		})

		It("uses extended input type", func() {
			prov := scheme.CreateConfigTypeSetConfigProvider()
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(prov.GetTypeOptionType(), TYPE),
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
			))
			cfg := Must(prov.GetConfigFor(opts))
			fmt.Printf("selected input options: %+v\n", cfg)

			spec := Must(scheme.GetInputSpecFor(cfg))
			Expect(spec).To(Equal(New("ghcr.io/open-component-model/image:v1.0")))
		})
	})
})

////////////////////////////////////////////////////////////////////////////////
// test input

const TYPE = "testinput"

type Spec struct {
	// PathSpec holds the repository path and tag of the image in the docker daemon
	cpi.PathSpec
}

var _ inputs.InputSpec = (*Spec)(nil)

func New(pathtag string) *Spec {
	return &Spec{
		PathSpec: cpi.NewPathSpec(TYPE, pathtag),
	}
}

func (s *Spec) Validate(fldPath *field.Path, ctx inputs.Context, inputFilePath string) field.ErrorList {
	allErrs := s.PathSpec.Validate(fldPath, ctx, inputFilePath)
	return allErrs
}

func (s *Spec) GetBlob(ctx inputs.Context, info inputs.InputResourceInfo) (blobaccess.BlobAccess, string, error) {
	blob := blobaccess.ForString(mime.MIME_TEXT, s.Path)
	return blob, "", nil
}

func ConfigHandler() flagsets.ConfigOptionTypeSetHandler {
	return cpi.NewMediaFileSpecOptionType(TYPE, AddConfig,
		options.PathOption, options.HintOption, options.PlatformsOption)
}

func AddConfig(opts flagsets.ConfigOptions, config flagsets.Config) error {
	if err := cpi.AddPathSpecConfig(opts, config); err != nil {
		return err
	}
	return nil
}
