// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/hpi"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/valuemergehandler/handlers/defaultmerge"
)

const PLUGIN = "merge"
const ALGORITHM = "acme.org/test"

var _ = Describe("plugin value merge handler", func() {
	var ctx ocm.Context
	var env *Builder
	var registry valuemergehandler.Registry

	BeforeEach(func() {
		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry = ctx.LabelMergeHandlers()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("executes handler", func() {
		registration.RegisterExtensions(ctx)

		Expect(registry.GetHandler(ALGORITHM)).NotTo(BeNil())

		spec := Must(valuemergehandler.NewSpecification(ALGORITHM, defaultmerge.NewConfig("test")))
		var local, inbound valuemergehandler.Value

		local.SetValue("local")
		inbound.SetValue("inbound")
		mod := Must(valuemergehandler.Merge(ctx, spec, "", local, &inbound))

		Expect(mod).To(BeTrue())
		Expect(inbound.RawMessage).To(YAMLEqual(`{"mode":"resolved"}`))
	})

	It("assigns specs", func() {
		registration.RegisterExtensions(ctx)

		Expect(registry.GetHandler(ALGORITHM)).NotTo(BeNil())

		s := ctx.LabelMergeHandlers().GetAssignment(hpi.LabelHint("testlabel", "v2"))
		Expect(s).NotTo(BeNil())
		Expect(s.Algorithm).To(Equal("simpleMapMerge"))
	})
})
