package uploaderoption

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	clictx "ocm.software/ocm/api/cli"
	"ocm.software/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
)

var _ = Describe("uploader option", func() {
	var flags *pflag.FlagSet
	var opt *Option
	var ctx clictx.Context

	BeforeEach(func() {
		ctx = clictx.New()
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
		opt = New(ctx.OCMContext())
		opt.AddFlags(flags)
	})

	It("parsed n:a:m", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:art:media={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "media",
			Config:       []byte(`{"name":"Name"}`),
		}}))
	})

	It("parsed n:a", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:art={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "",
			Config:       []byte(`{"name":"Name"}`),
		}}))
	})

	It("parsed n", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       []byte(`{"name":"Name"}`),
		}}))
	})

	It("parsed n::", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name::={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       []byte(`{"name":"Name"}`),
		}}))
	})

	It("parsed flat spec", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name=Name`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       []byte(`"Name"`),
		}}))
	})

	It("fails", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:::0:=Name`}))
		MustFailWithMessage(opt.Configure(ctx), "invalid uploader registration plugin/name:::0: (invalid priority) must be of "+optutils.RegistrationFormat)
	})
})
