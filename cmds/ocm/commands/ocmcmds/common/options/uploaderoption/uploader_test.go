// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package uploaderoption

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/spf13/pflag"

	"github.com/open-component-model/ocm/pkg/contexts/clictx"
)

var _ = Describe("uploader option", func() {
	var flags *pflag.FlagSet
	var opt *Option
	var ctx clictx.Context

	BeforeEach(func() {
		ctx = clictx.New()
		flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
		opt = &Option{}
		opt.AddFlags(flags)
	})

	It("parsed n:a:m", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:art:media={"name":"Name"}`}))
		MustBeSuccessful(opt.Complete(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "media",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n:a", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:art={"name":"Name"}`}))
		MustBeSuccessful(opt.Complete(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "art",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name={"name":"Name"}`}))
		MustBeSuccessful(opt.Complete(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed n::", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name::={"name":"Name"}`}))
		MustBeSuccessful(opt.Complete(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`{"name":"Name"}`),
		}}))
	})

	It("parsed flat spec", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name=Name`}))
		MustBeSuccessful(opt.Complete(ctx))

		Expect(opt.Registrations).To(Equal([]*Registration{{
			Name:         "plugin/name",
			ArtifactType: "",
			MediaType:    "",
			Config:       json.RawMessage(`"Name"`),
		}}))
	})

	It("fails", func() {
		MustBeSuccessful(flags.Parse([]string{`--uploader`, `plugin/name:::=Name`}))
		MustFailWithMessage(opt.Complete(ctx), "invalid uploader registration plugin/name::: must be of <name>[:<artifact type>[:<media type>]]")
	})
})
