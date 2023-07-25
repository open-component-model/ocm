// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package uploaderoption

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/v2/pkg/testutils"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/v2/cmds/ocm/commands/ocmcmds/common/options/optutils"
	"github.com/open-component-model/ocm/v2/pkg/contexts/clictx"
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
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n:a", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:art={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n::", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name::={"name":"Name"}`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed flat spec", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name=Name`}))
		MustBeSuccessful(opt.Configure(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`"Name"`),
		}}))
	})

	It("fails", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:::=Name`}))
		MustFailWithMessage(opt.Configure(ctx), "invalid uploader registration plugin/name::: must be of "+optutils.RegistrationFormat)
	})
})
