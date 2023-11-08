// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package composition_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	. "github.com/open-component-model/ocm/pkg/contexts/oci/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/relativeociref"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/comparch"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/composition"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/env/builder"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/refmgmt"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

const OCIPATH = "/tmp/oci"
const OCIHOST = "alias"

const RES = "ref"

var _ = Describe("cached access method blob", func() {
	var env *builder.Builder

	BeforeEach(func() {
		env = builder.NewBuilder()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("ocireg", func() {

		BeforeEach(func() {
			env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				OCIManifest1(env)
			})
			FakeOCIRepo(env, OCIPATH, OCIHOST)

			env.OCMCommonTransport(OCIPATH, accessio.FormatDirectory, func() {
				env.ComponentVersion(COMPONENT, VERSION, func() {
					env.Resource(RES, VERSION, "testtyp", v1.LocalRelation, func() {
						env.Access(relativeociref.New(OCINAMESPACE + ":" + OCIVERSION))
					})
				})
			})
		})

		It("caches blobs on close", func() {
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			srcfinalize := finalize.Nested()

			ctfrepo := Must(ctf.Open(env, accessobj.ACC_READONLY, OCIPATH, 0o700, env))

			refmgmt.AsLazy(ctfrepo)

			srcfinalize.Close(ctfrepo, "ctf")
			srccv := Must(ctfrepo.LookupComponentVersion(COMPONENT, VERSION))
			srcfinalize.Close(srccv, "src cv")

			res := Must(srccv.GetResource(v1.NewIdentity(RES)))
			srcblob := Must(res.BlobAccess())
			finalize.Close(srcblob, "source blob")
			Expect(srcblob.MimeType()).To(Equal("application/vnd.oci.image.manifest.v1+tar+gzip"))

			// now close the original environment
			// the blob access must be cached now and decoupled from the providing
			// repository.
			MustBeSuccessful(srcfinalize.Finalize())

			Expect(srcblob.MimeType()).To(Equal("application/vnd.oci.image.manifest.v1+tar+gzip"))
		})

		It("caches blobs on close", func() {
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			srcfinalize := finalize.Nested()

			ctfrepo := Must(ctf.Open(env, accessobj.ACC_READONLY, OCIPATH, 0o700, env))
			srcfinalize.Close(ctfrepo, "ctf")
			srccv := Must(ctfrepo.LookupComponentVersion(COMPONENT, VERSION))
			srcfinalize.Close(srccv, "src cv")

			res := Must(srccv.GetResource(v1.NewIdentity(RES)))

			// copy to composition repo
			repo := composition.NewRepository(env)
			finalize.Close(repo, "composition repository")
			MustBeSuccessful(repo.AddComponentVersion(srccv))

			// now close thenoriginal environment
			// the blob access must be cached now and decoupled from the providing
			// repository.
			MustBeSuccessful(srcfinalize.Finalize())

			cv := Must(repo.LookupComponentVersion(COMPONENT, VERSION))
			finalize.Close(cv, "composition cv")

			res = Must(cv.GetResource(v1.NewIdentity(RES)))

			m := Must(res.AccessMethod())
			finalize.Close(m, "copied method")

			Expect(m.MimeType()).To(Equal("application/vnd.oci.image.manifest.v1+tar+gzip"))
		})
	})

	Context("comparch", func() {
		BeforeEach(func() {
			env.ComponentArchive(OCIPATH, accessio.FormatTar, COMPONENT, VERSION, func() {
				env.Resource(RES, VERSION, "testtyp", v1.LocalRelation, func() {
					env.BlobStringData(mime.MIME_TEXT, "testdata")
				})
			})
		})

		It("caches blobs on close", func() {
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			srcfinalize := finalize.Nested()

			ctfrepo := Must(comparch.Open(env, accessobj.ACC_READONLY, OCIPATH, 0o700, env))
			srcfinalize.Close(ctfrepo, "ctf")
			srccv := ctfrepo

			res := Must(srccv.GetResource(v1.NewIdentity(RES)))
			srcblob := Must(res.BlobAccess())
			finalize.Close(srcblob, "source blob")
			Expect(srcblob.MimeType()).To(Equal(mime.MIME_TEXT))

			// now close the original environment
			// the blob access must be cached now and decoupled from the providing
			// repository.
			MustBeSuccessful(srcfinalize.Finalize())

			Expect(srcblob.MimeType()).To(Equal(mime.MIME_TEXT))
		})
	})
})
