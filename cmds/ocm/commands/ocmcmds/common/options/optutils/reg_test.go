package optutils_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/pkg/contexts/clictx"
	"github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("registration options", func() {
	var flags *pflag.FlagSet
	var opt optutils.RegistrationOption
	var ctx clictx.Context

	BeforeEach(func() {
		ctx = clictx.New()
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
		opt = optutils.NewRegistrationOption("test", "t", "test registration", `
- alice: some handler
- bob: some other handler
`)
		opt.AddFlags(flags)
	})

	It("parsed n:a:m", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name:art:media={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "media",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n:a", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name:art={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n::", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name::={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed flat spec", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name=Name`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`"Name"`),
		}}))
	})

	It("parsed config from file", func() {
		env := env.NewEnvironment(env.TestData())
		defer env.Cleanup()
		ctx := clictx.WithFileSystem(env).New()

		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name=@testdata/config`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*optutils.Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("fails", func() {
		MustBeSuccessful(flags.Parse([]string{`--test`, `plugin/name:::=Name`}))
		MustFailWithMessage(opt.Configure(ctx), "invalid test registration plugin/name::: must be of "+optutils.RegistrationFormat)
	})
})
