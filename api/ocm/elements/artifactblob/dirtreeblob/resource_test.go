package dirtreeblob_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testenv "github.com/open-component-model/ocm/api/helper/env"
	"github.com/open-component-model/ocm/api/ocm/compdesc"
	me "github.com/open-component-model/ocm/api/ocm/elements/artifactblob/dirtreeblob"
	"github.com/open-component-model/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/api/ocm/extensions/repositories/ctf"
	"github.com/open-component-model/ocm/api/ocm/extensions/resourcetypes"
	"github.com/open-component-model/ocm/api/utils/accessobj"
	"github.com/open-component-model/ocm/api/utils/mime"
	"github.com/open-component-model/ocm/api/utils/tarutils"
)

var _ = Describe("dir tree resource access", func() {
	var env *testenv.Environment

	BeforeEach(func() {
		env = testenv.NewEnvironment(testenv.TestData())
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("creates resource", func() {
		global := ociartifact.New("ghcr.io/mandelsoft/demo:v1.0.0")

		acc := me.ResourceAccess(env.OCMContext(), compdesc.NewResourceMeta("test", "", compdesc.LocalRelation), "testdata",
			me.WithExcludeFiles([]string{"dir/a"}),
			me.WithFileSystem(env.FileSystem()),
			me.WithHint("demo"),
			me.WithGlobalAccess(global),
		)

		Expect(acc.ReferenceHint()).To(Equal("demo"))
		Expect(acc.GlobalAccess()).To(Equal(global))
		Expect(acc.Meta().Type).To(Equal(resourcetypes.DIRECTORY_TREE))

		blob := Must(acc.BlobAccess())
		defer Defer(blob.Close, "blob")
		Expect(blob.MimeType()).To(Equal(mime.MIME_TAR))

		r := Must(blob.Reader())
		defer Defer(r.Close, "reader")
		files := Must(tarutils.ListArchiveContentFromReader(r))
		Expect(files).To(ConsistOf([]string{
			"dir",
			"dir/b",
			"dir/c",
		}))
	})

	It("adds resource", func() {
		global := ociartifact.New("ghcr.io/mandelsoft/demo:v1.0.0")

		acc := me.ResourceAccess(env.OCMContext(), compdesc.NewResourceMeta("test", "", compdesc.LocalRelation), "testdata",
			me.WithExcludeFiles([]string{"dir/a"}),
			me.WithFileSystem(env.FileSystem()),
			me.WithHint("demo"),
			me.WithGlobalAccess(global),
		)

		arch := Must(ctf.Create(env, accessobj.ACC_CREATE, "ctf", 0o700, env, accessobj.FormatDirectory))
		c := Must(arch.LookupComponent("arcme.org/test"))
		v := Must(c.NewVersion("v1.0.0"))

		MustBeSuccessful(v.SetResourceAccess(acc))
		MustBeSuccessful(c.AddVersion(v))
	})
})
