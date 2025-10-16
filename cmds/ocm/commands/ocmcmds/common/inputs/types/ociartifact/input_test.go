package ociartifact_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	. "ocm.software/ocm/api/oci/testhelper"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/cobrautils/flagsets"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/options"
	me "ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
	. "ocm.software/ocm/cmds/ocm/testhelper"
)

const (
	OCIPATH   = "/tmp/oci"
	OCIHOST   = "alias"
	ARCH      = "/tmp/ctf"
	VERSION   = "1.0.0"
	COMPONENT = "ocm.software/demo/test"
)

func CheckComponent(env *TestEnv) {
	repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
	defer Close(repo)
	cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
	defer Close(cv)
	cd := cv.GetDescriptor()

	Expect(string(cd.Provider.Name)).To(Equal("ocm.software"))

	r := Must(cv.GetResource(metav1.Identity{"name": "image"}))
	a := Must(r.Access())

	expDigest := "sha256:bde0f428596a33a6ba00b2df6047227e06130409fae69cf37edbe2eca13e8448"
	Expect(a.Describe(env.OCMContext())).To(Equal("Local blob " + expDigest + "[ocm.software/demo/test/image:v2.0-index]"))

	m := Must(r.AccessMethod())
	defer Close(m, "method")

	rd := Must(m.Reader())
	defer Close(rd, "reader")

	set := Must(artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(rd)))
	defer Close(set, "set")

	digest := set.GetMain()
	Expect(digest.Encoded()).To(Equal(D_OCIMANIFEST1))

	art := Must(set.GetArtifact(digest.String()))
	defer Close(art, "art")

	Expect(art.IsManifest()).To(BeTrue())
}

func Apply(opts flagsets.ConfigOptions) (inputs.InputSpec, error) {
	cfg := flagsets.Config{"type": me.TYPE}
	err := inputs.DefaultInputTypeScheme.GetInputType(me.TYPE).ConfigOptionTypeSetHandler().ApplyConfig(opts, cfg)
	if err != nil {
		return nil, err
	}
	fmt.Printf("config options: %+v\n", cfg)
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	return inputs.DefaultInputTypeScheme.Decode(data, nil)
}

var _ = Describe("Test Environment", func() {
	var (
		itype = inputs.DefaultInputTypeScheme.GetInputType(me.TYPE)
		flags *pflag.FlagSet
		opts  flagsets.ConfigOptions
		cfg   flagsets.Config
	)

	Context("options", func() {
		BeforeEach(func() {
			flags = &pflag.FlagSet{}
			opts = itype.ConfigOptionTypeSetHandler().CreateOptions()
			opts.AddFlags(flags)
			cfg = flagsets.Config{}
		})

		It("handles path option", func() {
			fmt.Printf("option names: %+v\n", opts.Names())

			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
			))
			MustBeSuccessful(itype.ConfigOptionTypeSetHandler().ApplyConfig(opts, cfg))

			spec := Must(Apply(opts))
			Expect(spec).To(Equal(me.New("ghcr.io/open-component-model/image:v1.0")))
		})

		It("handles platform option", func() {
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
				flagsets.OptionSpec(options.PlatformsOption, "linux/amd64"),
			))
			spec := Must(Apply(opts))
			Expect(spec).To(Equal(me.New("ghcr.io/open-component-model/image:v1.0", "linux/amd64")))
		})
	})

	Context("inputs", func() {
		BeforeEach(func() {
			flags = &pflag.FlagSet{}
			opts = inputs.DefaultInputTypeScheme.CreateOptions()
			opts.AddFlags(flags)
			cfg = flagsets.Config{}
		})

		It("input type", func() {
			fmt.Printf("input option names: %+v\n", opts.Names())
			MustBeSuccessful(flagsets.ParseOptionsFor(flags,
				flagsets.OptionSpec(inputs.DefaultInputTypeScheme.ConfigTypeSetConfigProvider().GetTypeOptionType(), me.TYPE),
				flagsets.OptionSpec(options.PathOption, "ghcr.io/open-component-model/image:v1.0"),
				flagsets.OptionSpec(options.PlatformsOption, "linux/amd64"),
				flagsets.OptionSpec(options.PlatformsOption, "/arm64"),
			))
			cfg := Must(inputs.DefaultInputTypeScheme.GetConfigFor(opts))
			fmt.Printf("selected input options: %+v\n", cfg)

			spec := Must(inputs.DefaultInputTypeScheme.GetInputSpecFor(opts))
			Expect(spec).To(Equal(me.New("ghcr.io/open-component-model/image:v1.0", "linux/amd64", "/arm64")))
		})
	})

	Context("scenario", func() {
		var env *TestEnv
		var rname string

		BeforeEach(func() {
			env = NewTestEnv(TestData())

			rname = FakeOCIRepo(env.Builder, OCIPATH, OCIHOST)

			fmt.Printf("image url: %s\n", oci.StandardOCIRef(rname, OCINAMESPACE3, OCIINDEXVERSION))
			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIIndex1(env.Builder)
			})
		})

		AfterEach(func() {
			env.Cleanup()
		})

		It("creates ctf and adds component", func() {
			Expect(env.Execute("add", "c", "-fc", "--file", ARCH, "testdata/component-constructor.yaml")).To(Succeed())
			Expect(env.DirExists(ARCH)).To(BeTrue())
			CheckComponent(env)
		})
	})
})
