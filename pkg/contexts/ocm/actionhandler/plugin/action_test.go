// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package plugin_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/registration"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/datacontext/action/handlers"
	oci_repository_prepare "github.com/open-component-model/ocm/pkg/contexts/datacontext/action/types/oci-repository-prepare"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/actionhandler/plugin"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugincacheattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/plugindirattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/plugin/plugins"
)

const PLUGIN = "test"

var _ = Describe("plugin action handler", func() {
	var ctx ocm.Context
	var registry plugins.Set
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder(nil)
		ctx = env.OCMContext()
		plugindirattr.Set(ctx, "testdata")
		registry = plugincacheattr.Get(ctx)
		p := registry.Get("action")
		Expect(p).NotTo(BeNil())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("executes with no plugin registration", func() {
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "ghcr.io", "mandelsoft", nil))
		Expect(result).To(BeNil())
	})

	It("executes with no handler", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "mandelsoft.org", "mandelsoft", nil))
		Expect(result).To(BeNil())
	})

	It("used default registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction("test"), handlers.ForSelectors("mandelsoft.org"))
		MustFailWithMessage(plugin.RegisterActionHandler(ctx.AttributesContext(), "action", ctx, opts),
			"action \"test\" is unknown for plugin action")
	})

	It("uses default registration", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "ghcr.io", "mandelsoft", nil))
		Expect(result).NotTo(BeNil())
		Expect(result.Message).To(Equal("all good"))
	})

	FIt("uses default pattern registration", func() {
		registration.RegisterExtensions(ctx)
		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "xyz.dkr.ecr.us-west-2.amazonaws.com", "mandelsoft", nil))
		Expect(result).NotTo(BeNil())
		Expect(result.Message).To(Equal("all good"))
	})

	It("executes action for dynamic registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction(oci_repository_prepare.Type), handlers.ForSelectors("mandelsoft.org"))
		MustBeSuccessful(plugin.RegisterActionHandler(ctx.AttributesContext(), "action", ctx, opts))

		result := Must(oci_repository_prepare.Execute(ctx.GetActions(), "mandelsoft.org", "mandelsoft", nil))
		Expect(result.Message).To(Equal("all good"))
	})

	It("executed action after abstract registration", func() {
		registration.RegisterExtensions(ctx)
		opts := handlers.NewOptions(handlers.ForAction(oci_repository_prepare.Type), handlers.ForAction(oci_repository_prepare.Type), handlers.ForSelectors("mandelsoft.org"))
		ok := Must(ctx.GetActions().RegisterByName("plugin/action", ctx.OCIContext(), ctx, opts))
		Expect(ok).To(BeTrue())

	})
	/*
		It("uploads after abstract registration", func() {
			repo := Must(ctf.Open(ctx, accessobj.ACC_READONLY, ARCH, 0, env))
			defer Close(repo, "source repo")

			cv := Must(repo.LookupComponentVersion(COMP, VERS))
			defer Close(cv, "source version")

			MustFailWithMessage(registration.RegisterBlobHandlerByName(ctx, "plugin/test", []byte("{}"), registration.ForArtifactType(RSCTYPE)),
				//MustFailWithMessage(plugin.RegisterBlobHandler(env.OCMContext(), "test", "", RSCTYPE, "", []byte("{}")),
				"plugin uploader test/testuploader: path missing in repository spec",
			)
			repospec := Must(json.Marshal(repoSpec))
			MustBeSuccessful(registration.RegisterBlobHandlerByName(ctx, "plugin/test", repospec))

			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0700, accessio.FormatDirectory, env))
			defer Close(tgt, "target repo")

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, Must(standard.New(standard.ResourcesByValue()))))
			Expect(env.DirExists(OUT)).To(BeTrue())

			Expect(vfs.FileExists(osfs.New(), filepath.Join(repodir, REPO, HINT))).To(BeTrue())

			tcv := Must(tgt.LookupComponentVersion(COMP, VERS))
			defer Close(tcv, "target version")

			r := Must(tcv.GetResourceByIndex(0))
			a := Must(r.Access())

			var spec AccessSpec
			MustBeSuccessful(json.Unmarshal(Must(json.Marshal(a)), &spec))
			Expect(spec).To(Equal(*accessSpec))

			m := Must(a.AccessMethod(tcv))
			defer Close(m, "method")

			Expect(string(Must(m.Get()))).To(Equal(CONTENT))
		})

	*/
})
