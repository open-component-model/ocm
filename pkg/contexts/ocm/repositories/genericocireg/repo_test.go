// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package genericocireg_test

import (
	"path"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/open-component-model/ocm/pkg/common/accessio"
	"github.com/open-component-model/ocm/pkg/common/accessobj"
	"github.com/open-component-model/ocm/pkg/contexts/oci"
	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/artifactset"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf"
	"github.com/open-component-model/ocm/pkg/contexts/oci/repositories/ctf/testhelper"
	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/attrs/compatattr"
	storagecontext "github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/blobhandler/oci/ocirepo"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/repositories/genericocireg/componentmapping"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/resourcetypes"
	ocmutils "github.com/open-component-model/ocm/pkg/contexts/ocm/utils"
	"github.com/open-component-model/ocm/pkg/finalizer"
	"github.com/open-component-model/ocm/pkg/mime"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var DefaultContext = ocm.New()

const COMPONENT = "github.com/mandelsoft/ocm"
const TESTBASE = "testbase.de"

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		ocispec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
		spec = genericocireg.NewRepositorySpec(ocispec, nil)

	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("creates a dummy component", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		MustBeSuccessful(comp.AddVersion(vers))

		MustBeSuccessful(finalize.Finalize())

		// access it again
		repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))

		ok := Must(repo.ExistsComponentVersion(COMPONENT, "v1"))
		Expect(ok).To(BeTrue())

		comp = finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers = finalizer.ClosingWith(&finalize, Must(comp.LookupVersion("v1")))
		Expect(vers.GetDescriptor()).To(Equal(compdesc.New(COMPONENT, "v1")))

		MustBeSuccessful(finalize.Finalize())
	})

	It("handles legacylocalociblob  access method", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		blob := accessio.BlobAccessForString(mime.MIME_OCTET, "anydata")

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		acc := Must(vers.AddBlob(blob, "", "", nil))

		// check provided actual access to be local blob
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		l, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).To(BeNil())

		acc = localociblob.New(digest.Digest(l.LocalReference))

		MustBeSuccessful(vers.SetResource(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))

		rs := Must(vers.GetResourceByIndex(0))

		spec := Must(rs.Access())
		Expect(spec.GetType()).To(Equal(localociblob.Type))

		m := Must(rs.AccessMethod())
		finalize.Close(m)
		Expect(m.MimeType()).To(Equal("application/octet-stream"))
		data := Must(m.Get())
		Expect(string(data)).To(Equal("anydata"))
	})

	It("imports blobs", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewBlobHandler(base))).New()
		blob := accessio.BlobAccessForString(mime.MIME_OCTET, "anydata")

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		acc := Must(vers.AddBlob(blob, "", "", nil))

		// check provided actual access to be local blob
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		l, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).NotTo(BeNil())

		// check provided global access to be oci blob
		g := Must(l.GlobalAccess.Evaluate(DefaultContext))
		o, ok := g.(*ociblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(o.Digest).To(Equal(blob.Digest()))
		Expect(o.Reference).To(Equal(TESTBASE + "/" + componentmapping.ComponentDescriptorNamespace + "/" + COMPONENT))
		MustBeSuccessful(vers.SetResource(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, v1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))
	})

	It("imports artifact", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		mime := artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest) + "+tar+gzip"
		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(ocirepo.NewArtifactHandler(base), cpi.ForMimeType(mime))).New()

		// create artifactset
		opts := Must(accessio.AccessOptions(nil, accessio.PathFileSystem(tempfs)))
		r := Must(artifactset.FormatTGZ.Create("test.tgz", opts, 0700))
		testhelper.DefaultManifestFill(r)
		r.Annotate(artifactset.MAINARTIFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		Expect(r.Close()).To(Succeed())

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
		ocirepo := repo.(*genericocireg.Repository).GetOCIRepository()
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		nested := finalize.Nested()
		comp := finalizer.ClosingWith(nested, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(nested, Must(comp.NewVersion("v1")))
		blob := accessio.BlobAccessForFile(mime, "test.tgz", tempfs)

		acc := Must(vers.AddBlob(blob, "", "artifact1", nil))
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o := acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact1@sha256:" + testhelper.DIGEST_MANIFEST))
		MustBeSuccessful(comp.AddVersion(vers))

		acc = Must(vers.AddBlob(blob, "", "artifact2:v1", nil))
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o = acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact2:v1"))
		MustBeSuccessful(comp.AddVersion(vers))

		MustBeSuccessful(nested.Finalize())

		ns := finalizer.ClosingWith(nested, Must(ocirepo.LookupNamespace("artifact2")))
		art := finalizer.ClosingWith(nested, Must(ns.GetArtifact("v1")))
		testhelper.CheckArtifact(art)

		MustBeSuccessful(finalize.Finalize())
	})

	It("removes old unused layers", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize, "finalize open elements")

		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		Expect(reflect.TypeOf(repo).String()).To(Equal("*genericocireg.Repository"))

		nested := finalize.Nested()

		comp := finalizer.ClosingWith(nested, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(nested, Must(comp.NewVersion("v1")))

		m1 := compdesc.NewResourceMeta("rsc1", resourcetypes.PLAIN_TEXT, v1.LocalRelation)
		blob := accessio.BlobAccessForString(mime.MIME_TEXT, "testdata")

		MustBeSuccessful(vers.SetResourceBlob(m1, blob, "", nil))
		MustBeSuccessful(comp.AddVersion(vers))

		MustBeSuccessful(nested.Finalize())

		// modify rsource in component
		vers = finalizer.ClosingWith(nested, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
		blob = accessio.BlobAccessForString(mime.MIME_TEXT, "otherdata")
		MustBeSuccessful(vers.SetResourceBlob(m1, blob, "", nil))
		MustBeSuccessful(nested.Finalize())

		// check content
		vers = finalizer.ClosingWith(nested, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
		r := Must(vers.GetResource(v1.NewIdentity("rsc1")))
		data := Must(ocmutils.GetResourceData(r))
		Expect(string(data)).To(Equal("otherdata"))
		MustBeSuccessful(nested.Finalize())

		MustBeSuccessful(finalize.Finalize())

		ocirepo := Must(DefaultContext.OCIContext().RepositoryForSpec(ocispec))
		finalize.Close(ocirepo)

		art := Must(ocirepo.LookupArtifact("component-descriptors/"+COMPONENT, "v1"))
		finalize.Close(art)

		Expect(art.GetDescriptor().IsManifest()).To(BeTrue())
		Expect(len(art.GetDescriptor().Manifest().Layers)).To(Equal(2))
	})

	Context("legacy mode", func() {
		It("creates a legacy dummy component", func() {
			ctx := ocm.New()
			compatattr.Set(ctx, true)

			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
			comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
			vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
			MustBeSuccessful(comp.AddVersion(vers))
			MustBeSuccessful(finalize.Finalize())

			// access as OCI repository

			ocirepo := finalizer.ClosingWith(&finalize, Must(oci.DefaultContext().RepositoryForSpec(ocispec)))
			ns := finalizer.ClosingWith(&finalize, Must(ocirepo.LookupNamespace(path.Join(componentmapping.ComponentDescriptorNamespace, COMPONENT))))
			art := finalizer.ClosingWith(&finalize, Must(ns.GetArtifact("v1")))
			m := art.GetDescriptor().Manifest()
			Expect(m.Config.MediaType).To(Equal(componentmapping.LegacyComponentDescriptorConfigMimeType))
			Expect(len(m.Layers)).To(Equal(1))
			Expect(m.Layers[0].MediaType).To(Equal(componentmapping.LegacyComponentDescriptorTarMimeType))
		})

		It("updates a legacy dummy component", func() {
			ctx := ocm.New()
			compatattr.Set(ctx, true)

			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
			comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
			vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
			MustBeSuccessful(comp.AddVersion(vers))
			MustBeSuccessful(finalize.Finalize())

			// update component in non-legacy context

			repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
			comp = finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
			vers = finalizer.ClosingWith(&finalize, Must(comp.LookupVersion("v1")))
			vers.GetDescriptor().Provider.Name = "acme.org"
			MustBeSuccessful(comp.AddVersion(vers))
			MustBeSuccessful(finalize.Finalize())

			// access as OCI repository

			ocirepo := finalizer.ClosingWith(&finalize, Must(oci.DefaultContext().RepositoryForSpec(ocispec)))
			ns := finalizer.ClosingWith(&finalize, Must(ocirepo.LookupNamespace(path.Join(componentmapping.ComponentDescriptorNamespace, COMPONENT))))
			art := finalizer.ClosingWith(&finalize, Must(ns.GetArtifact("v1")))
			m := art.GetDescriptor().Manifest()
			Expect(m.Config.MediaType).To(Equal(componentmapping.LegacyComponentDescriptorConfigMimeType))
			Expect(len(m.Layers)).To(Equal(1))
			Expect(m.Layers[0].MediaType).To(Equal(componentmapping.LegacyComponentDescriptorTarMimeType))
			MustBeSuccessful(finalize.Finalize())

			repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
			vers = finalizer.ClosingWith(&finalize, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
			Expect(string(vers.GetDescriptor().Provider.Name)).To(Equal("acme.org"))
		})
	})
})
