package helmblob_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/helper/env"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	me "ocm.software/ocm/api/ocm/elements/artifactblob/helmblob"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessobj"
)

var _ = Describe("", func() {
	var e *Builder

	BeforeEach(func() {
		e = NewBuilder(env.TestData())
	})

	AfterEach(func() {
		MustBeSuccessful(e.Cleanup())
	})

	It("", func() {
		ctf := Must(ctfocm.Open(e, accessobj.ACC_CREATE, "/repo", 0o700, e, ctfocm.FormatDirectory))
		defer Close(ctf)
		cv := Must(ctf.NewComponentVersion("ocm.software/test-component", "1.0.0"))
		defer Close(cv)
		MustBeSuccessful(cv.SetResourceByAccess(me.ResourceAccess(e.OCMContext(), cpi.NewResourceMeta("helm1", "blob", metav1.LocalRelation), "/testdata/testchart1", me.WithFileSystem(e.FileSystem()))))
		MustBeSuccessful(cv.SetResourceByAccess(me.ResourceAccess(e.OCMContext(), cpi.NewResourceMeta("helm2", "blob", metav1.LocalRelation), "/testdata/testchart2", me.WithFileSystem(e.FileSystem()))))
		MustBeSuccessful(ctf.AddComponentVersion(cv, true))
	})
})
