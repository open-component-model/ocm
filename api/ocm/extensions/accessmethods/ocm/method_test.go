package ocm_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ocm"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH  = "/tmp/ctf"
	COMP1 = "acme.org/test1"
	COMP2 = "acme.org/test2"
	VERS  = "v1"
	RSC1  = "resource1"
	RSC2  = "resource2"
	REF1  = "reference1"

	DATA = "some test data"
)

var _ = Describe("Method", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("remote access", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP1, VERS, func() {
					env.Resource(RSC1, VERS, resourcetypes.PLAIN_TEXT, v1.LocalRelation, func() {
						env.ExtraIdentity("other", "value")
						env.BlobStringData(mime.MIME_TEXT, DATA)
					})
				})

				env.ComponentVersion(COMP2, VERS, func() {
					env.Reference(REF1, COMP1, VERS, func() {
						env.ExtraIdentity("purpose", "test")
					})
				})
			})
		})

		It("accesses artifact", func() {
			spec := Must(ocm.New(COMP1, VERS, Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, env)), v1.NewIdentity(RSC1, "other", "value")))

			m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()}))
			defer Close(m)
			data := Must(m.Get())
			Expect(string(data)).To(Equal(DATA))
		})

		It("accesses indirect artifact", func() {
			spec := Must(ocm.New(COMP2, VERS, Must(ctf.NewRepositorySpec(accessobj.ACC_READONLY, ARCH, env)), v1.NewIdentity(RSC1, "other", "value"), v1.NewIdentity(REF1, "purpose", "test")))

			m := Must(spec.AccessMethod(&cpi.DummyComponentVersionAccess{env.OCMContext()}))
			defer Close(m)
			data := Must(m.Get())
			Expect(string(data)).To(Equal(DATA))
		})
	})

	Context("local access", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMP1, VERS, func() {
					env.Resource(RSC1, VERS, resourcetypes.PLAIN_TEXT, v1.LocalRelation, func() {
						env.ExtraIdentity("other", "value")
						env.BlobStringData(mime.MIME_TEXT, DATA)
					})
				})

				env.ComponentVersion(COMP2, VERS, func() {
					env.Resource(RSC2, VERS, resourcetypes.PLAIN_TEXT, v1.LocalRelation, func() {
						env.Access(Must(ocm.New(COMP1, VERS, nil, v1.NewIdentity(RSC1, "other", "value"))))
					})
				})
			})
		})

		It("accesses artifact", func() {
			repo := Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
			defer Close(repo, "repo")

			cv := Must(repo.LookupComponentVersion(COMP2, VERS))
			defer Close(cv, "cv")

			ra := Must(cv.GetResourceByIndex(0))

			m := Must(ra.AccessMethod())
			defer Close(m)
			data := Must(m.Get())
			Expect(string(data)).To(Equal(DATA))
		})
	})
})
