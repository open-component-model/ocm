package virtual_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/env"
	. "github.com/open-component-model/ocm/pkg/env/builder"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"

	"github.com/open-component-model/ocm/pkg/blobaccess"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/compose"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compositionmodeattr"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/virtual"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/virtual/example"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
)

var _ = Describe("virtual repo", func() {
	var env *Builder
	var repo ocm.Repository
	var access *example.Access

	// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

	AfterEach(func() {
		MustBeSuccessful(repo.Close())
		env.Cleanup()
	})

	Context("readonly", func() {
		BeforeEach(func() {
			env = NewBuilder(TestData())
			access = Must(example.NewAccess(Must(projectionfs.New(env, "testdata")), true))
			repo = virtual.NewRepository(env.OCMContext(), access)
		})

		It("handles list", func() {
			lister := repo.ComponentLister()
			Expect(lister).NotTo(BeNil())
			names := Must(lister.GetComponents("", true))
			Expect(names).To(ConsistOf([]string{"acme.org/component", "acme.org/component/ref"}))
		})

		It("handles get", func() {
			comp := Must(repo.LookupComponent("acme.org/component"))
			defer Close(comp, "component")
			Expect(comp.ListVersions()).To(ConsistOf([]string{"v1.0.0"}))
			Expect(comp.HasVersion("v1.0.0")).To(BeTrue())
			Expect(comp.HasVersion("v1.0.1")).To(BeFalse())
			vers := Must(comp.LookupVersion("v1.0.0"))
			defer Close(vers, "version")
			r := Must(vers.GetResourceByIndex(0))
			data := Must(ocmutils.GetResourceData(r))
			Expect(string(data)).To(Equal("my test data\n"))

			a := Must(r.Access())
			Expect(a.GetInexpensiveContentVersionIdentity(vers)).To(Equal("sha256:2fdeb101f225dad71efd2dadb92b5aa422169f1884eecb81abdd988d77b68466"))
		})
	})

	Context("modifiable", func() {
		BeforeEach(func() {
			env = NewBuilder(TestData())

			fs := Must(projectionfs.New(env, "testdata"))
			fs = layerfs.New(memoryfs.New(), fs)
			access = Must(example.NewAccess(fs, false))
			repo = virtual.NewRepository(env.OCMContext(), access)
		})

		DescribeTable("handles put", func(mode bool, typ string) {
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			compositionmodeattr.Set(env.OCMContext(), mode)
			comp := Must(repo.LookupComponent("acme.org/component/new"))
			finalize.Close(comp, "component")
			Expect(comp.ListVersions()).To(ConsistOf([]string{}))
			vers := Must(comp.NewVersion("v1.0.0", false))
			finalize.Close(vers, "version")

			blob := blobaccess.ForString(mime.MIME_TEXT, "new test data")
			MustBeSuccessful(vers.SetResourceBlob(compdesc.NewResourceMeta("new", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), blob, "", nil))

			r := Must(vers.GetResourceByIndex(0))
			a := Must(r.Access())
			Expect(a.GetKind()).To(Equal(typ))

			comp.AddVersion(vers)
			r = Must(vers.GetResourceByIndex(0)) // re-read resource from component descriptor.
			a = Must(r.Access())
			Expect(a.GetKind()).To(Equal(localblob.Type))

			dig := "fe81d80611e39a10f1d7d12f98ce0bc6fe745d08fef007d8eebddc0a21d17827"
			Expect(a.(*localblob.AccessSpec).LocalReference).To(Equal(dig))

			MustBeSuccessful(finalize.Finalize())

			MustBeSuccessful(access.Reset())

			comp = Must(repo.LookupComponent("acme.org/component/new"))
			finalize.Close(comp, "component")
			Expect(comp.ListVersions()).To(ConsistOf([]string{"v1.0.0"}))
			Expect(comp.HasVersion("v1.0.0")).To(BeTrue())
			Expect(comp.HasVersion("v1.0.1")).To(BeFalse())
			vers = Must(comp.LookupVersion("v1.0.0"))
			finalize.Close(vers, "version")
			r = Must(vers.GetResourceByIndex(0))
			data := Must(ocmutils.GetResourceData(r))
			Expect(string(data)).To(Equal("new test data"))

			a = Must(r.Access())
			Expect(a.GetInexpensiveContentVersionIdentity(vers)).To(Equal("sha256:" + dig))

		},
			Entry("with direct mode", false, localblob.Type),
			Entry("with composition mode", true, compose.Type),
		)
	})
})
