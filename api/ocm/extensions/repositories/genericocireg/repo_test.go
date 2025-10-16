package genericocireg_test

import (
	"fmt"
	"io"
	"path"
	"reflect"

	"github.com/mandelsoft/goutils/finalizer"
	. "github.com/mandelsoft/goutils/testutils"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	ocicpi "ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf/testhelper"
	"ocm.software/ocm/api/oci/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/cpi/repocpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localociblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociblob"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/compatattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/keepblobattr"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	handler "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo"
	"ocm.software/ocm/api/ocm/extensions/digester/digesters/artifact"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg/componentmapping"
	ocmreg "ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
	"ocm.software/ocm/api/ocm/ocmutils"
	ocmtesthelper "ocm.software/ocm/api/ocm/testhelper"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

var DefaultContext = ocm.New()

const (
	COMPONENT = "github.com/mandelsoft/ocm"
	TESTBASE  = "testbase.de"
)

var _ = Describe("component repository mapping", func() {
	var tempfs vfs.FileSystem

	var ocispec oci.RepositorySpec
	var spec *genericocireg.RepositorySpec

	BeforeEach(func() {
		t, err := osfs.NewTempFileSystem()
		Expect(err).To(Succeed())
		tempfs = t

		// ocmlog.Context().AddRule(logging.NewConditionRule(logging.TraceLevel, accessio.ALLOC_REALM))

		ocispec, err = ctf.NewRepositorySpec(accessobj.ACC_CREATE, "test", accessio.PathFileSystem(tempfs), accessobj.FormatDirectory)
		Expect(err).To(Succeed())
		spec = genericocireg.NewRepositorySpec(ocispec, nil)
	})

	AfterEach(func() {
		vfs.Cleanup(tempfs)
	})

	It("Don't Panic! When it's not a semver.org conform version. #756", func() {
		repo := Must(DefaultContext.RepositoryForSpec(spec))
		comp := Must(repo.LookupComponent(COMPONENT))
		cva, err := comp.NewVersion("v1.two.zeo-2")
		Expect(err).To(HaveOccurred())
		Expect(cva).To(BeNil())
		Expect(err.Error()).To(Equal("invalid semantic version"))
	})

	It("creates a dummy component", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		MustBeSuccessful(comp.AddVersion(vers))

		noref := vers.Repository()
		Expect(noref).NotTo(BeNil())
		Expect(noref.IsClosed()).To(BeFalse())
		Expect(noref.Close()).To(Succeed())
		Expect(noref.IsClosed()).To(BeFalse())

		MustBeSuccessful(finalize.Finalize())

		Expect(noref.IsClosed()).To(BeTrue())
		Expect(noref.Close()).To(MatchError("closed"))
		ExpectError(noref.LookupComponent("dummy")).To(MatchError("closed"))

		// access it again
		repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))

		ok := Must(repo.ExistsComponentVersion(COMPONENT, "v1"))
		Expect(ok).To(BeTrue())

		comp = finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers = finalizer.ClosingWith(&finalize, Must(comp.LookupVersion("v1")))
		Expect(vers.GetDescriptor()).To(Equal(compdesc.New(COMPONENT, "v1")))

		MustBeSuccessful(finalize.Finalize())
	})

	const ref4 = "sha256:9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08,sha256:3a6eb0790f39ac87c94f3856b2dd2c5d110e6811602261a9a923d3bb23adc8b7"
	const ref5 = "sha256:a4853613b2a38568ed4e49196238152469097412d06d5e5fc9be8ab92cfdf2bf,sha256:977817f6f61f4dd501df3036a3e16b31452b36f4aa3edcf9a3f3242a79d7170d"
	const ref8 = "sha256:" + ocmtesthelper.D_TESTDATA

	DescribeTable("creates a dummy component with chunks", func(limit int, f func(ocm.ResourceAccess), ref string) {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))
		repocpi.SetBlobLimit(impl, int64(limit))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))

		m1 := compdesc.NewResourceMeta("rsc1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		blob := blobaccess.ForString(mime.MIME_TEXT, ocmtesthelper.S_TESTDATA)
		MustBeSuccessful(vers.SetResourceBlob(m1, blob, "", nil))

		MustBeSuccessful(comp.AddVersion(vers))

		noref := vers.Repository()
		Expect(noref).NotTo(BeNil())
		Expect(noref.IsClosed()).To(BeFalse())
		Expect(noref.Close()).To(Succeed())
		Expect(noref.IsClosed()).To(BeFalse())

		MustBeSuccessful(finalize.Finalize())

		Expect(noref.IsClosed()).To(BeTrue())
		Expect(noref.Close()).To(MatchError("closed"))
		ExpectError(noref.LookupComponent("dummy")).To(MatchError("closed"))

		// access it again
		repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))

		ok := Must(repo.ExistsComponentVersion(COMPONENT, "v1"))
		Expect(ok).To(BeTrue())

		comp = finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers = finalizer.ClosingWith(&finalize, Must(comp.LookupVersion("v1")))

		rsc := Must(vers.GetResourceByIndex(0))
		acc := Must(rsc.Access())
		local, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		// fmt.Printf("localref: %s\n", local.LocalReference)
		Expect(local.LocalReference).To(Equal(ref))
		Expect(rsc.Meta().Digest).NotTo(BeNil())
		Expect(rsc.Meta().Digest.Value).To(Equal(ocmtesthelper.D_TESTDATA))

		f(rsc)

		MustBeSuccessful(finalize.Finalize())
	},
		Entry("get blob", 5, func(rsc ocm.ResourceAccess) {
			data := Must(ocmutils.GetResourceData(rsc))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref5),
		Entry("stream blob", 5, func(rsc ocm.ResourceAccess) {
			r := Must(ocmutils.GetResourceReader(rsc))
			data := Must(io.ReadAll(r))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref5),
		Entry("stream blob with small buffer", 5, func(rsc ocm.ResourceAccess) {
			var buf [2]byte
			var data []byte

			r := Must(ocmutils.GetResourceReader(rsc))

			for {
				n, err := r.Read(buf[:])
				if n > 0 {
					data = append(data, buf[:n]...)
				}
				if err != nil {
					if err == io.EOF {
						break
					} else {
						MustBeSuccessful(err)
					}
				}
			}
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref5),

		Entry("get blob (match limit)", len(ocmtesthelper.S_TESTDATA), func(rsc ocm.ResourceAccess) {
			data := Must(ocmutils.GetResourceData(rsc))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref8),
		Entry("stream blob (match limit)", len(ocmtesthelper.S_TESTDATA), func(rsc ocm.ResourceAccess) {
			r := Must(ocmutils.GetResourceReader(rsc))
			data := Must(io.ReadAll(r))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref8),
		Entry("stream blob with small buffer  (match limit)", len(ocmtesthelper.S_TESTDATA), func(rsc ocm.ResourceAccess) {
			var buf [2]byte
			var data []byte

			r := Must(ocmutils.GetResourceReader(rsc))

			for {
				n, err := r.Read(buf[:])
				if n > 0 {
					data = append(data, buf[:n]...)
				}
				if err != nil {
					if err == io.EOF {
						break
					} else {
						MustBeSuccessful(err)
					}
				}
			}
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref8),

		Entry("get blob (match limit/2)", len(ocmtesthelper.S_TESTDATA)/2, func(rsc ocm.ResourceAccess) {
			data := Must(ocmutils.GetResourceData(rsc))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref4),
		Entry("stream blob (match limit/2)", len(ocmtesthelper.S_TESTDATA)/2, func(rsc ocm.ResourceAccess) {
			r := Must(ocmutils.GetResourceReader(rsc))
			data := Must(io.ReadAll(r))
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref4),
		Entry("stream blob with small buffer  (match limit/2)", len(ocmtesthelper.S_TESTDATA)/2, func(rsc ocm.ResourceAccess) {
			var buf [2]byte
			var data []byte

			r := Must(ocmutils.GetResourceReader(rsc))

			for {
				n, err := r.Read(buf[:])
				if n > 0 {
					data = append(data, buf[:n]...)
				}
				if err != nil {
					if err == io.EOF {
						break
					} else {
						MustBeSuccessful(err)
					}
				}
			}
			Expect(string(data)).To(Equal(ocmtesthelper.S_TESTDATA))
		}, ref4),
	)

	It("handles legacylocalociblob access method", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		blob := blobaccess.ForString(mime.MIME_OCTET, "anydata")

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		acc := Must(vers.AddBlob(blob, "", "", nil))

		MustBeSuccessful(vers.SetResource(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))

		rs := Must(vers.GetResourceByIndex(0))
		acc = Must(rs.Access())

		// check provided actual access to be local blob
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		l, ok := acc.(*localblob.AccessSpec)
		Expect(ok).To(BeTrue())
		Expect(l.LocalReference).To(Equal(blob.Digest().String()))
		Expect(l.GlobalAccess).To(BeNil())

		acc = localociblob.New(digest.Digest(l.LocalReference))

		MustBeSuccessful(vers.SetResource(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))

		rs = Must(vers.GetResourceByIndex(0))
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
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(handler.NewBlobHandler(base))).New()
		blob := blobaccess.ForString(mime.MIME_OCTET, ocmtesthelper.S_TESTDATA)

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))

		comp := finalizer.ClosingWith(&finalize, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(&finalize, Must(comp.NewVersion("v1")))
		acc := Must(vers.AddBlob(blob, "", "", nil))
		MustBeSuccessful(vers.SetResource(compdesc.NewResourceMeta("blob", resourcetypes.PLAIN_TEXT, metav1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))

		res := Must(vers.GetResourceByIndex(0))
		acc = Must(res.Access())

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

		Expect(res.Meta().Digest).NotTo(BeNil())
		Expect(res.Meta().Digest.Value).To(Equal(ocmtesthelper.D_TESTDATA))
	})

	It("imports artifact", func() {
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		mime := artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest) + "+tar+gzip"
		base := func(ctx *storagecontext.StorageContext) string {
			return TESTBASE
		}
		ctx := ocm.WithBlobHandlers(ocm.DefaultBlobHandlers().Copy().Register(handler.NewArtifactHandler(base), cpi.ForMimeType(mime))).New()
		keepblobattr.Set(ctx, true)

		// create artifactset
		opts := Must(accessio.AccessOptions(nil, accessio.PathFileSystem(tempfs)))
		r := Must(artifactset.FormatTGZ.Create("test.tgz", opts, 0o700))
		testhelper.DefaultManifestFill(r)
		r.Annotate(artifactset.MAINARTIFACT_ANNOTATION, "sha256:"+testhelper.DIGEST_MANIFEST)
		Expect(r.Close()).To(Succeed())

		// create repository
		repo := finalizer.ClosingWith(&finalize, Must(ctx.RepositoryForSpec(spec)))
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))
		ocirepo := genericocireg.GetOCIRepository(repo)
		Expect(ocirepo).NotTo(BeNil())

		nested := finalize.Nested()
		comp := finalizer.ClosingWith(nested, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(nested, Must(comp.NewVersion("v1")))
		blob := blobaccess.ForFile(mime, "test.tgz", tempfs)

		fmt.Printf("physical digest: %s\n", blob.Digest())
		acc := Must(vers.AddBlob(blob, "", "artifact1", nil))
		MustBeSuccessful(vers.SetResource(cpi.NewResourceMeta("image", resourcetypes.OCI_IMAGE, metav1.LocalRelation), acc))
		MustBeSuccessful(comp.AddVersion(vers))

		res := Must(vers.GetResourceByIndex(0))
		acc = Must(res.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))
		rd := res.Meta().Digest
		Expect(rd).NotTo(BeNil())
		Expect(rd.Value).To(Equal(testhelper.DIGEST_MANIFEST))
		Expect(rd.NormalisationAlgorithm).To(Equal(artifact.OciArtifactDigestV1))
		Expect(rd.HashAlgorithm).To(Equal(sha256.Algorithm))

		acc = acc.GlobalAccessSpec(ctx)
		Expect(acc).NotTo(BeNil())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o := acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact1@sha256:" + testhelper.DIGEST_MANIFEST))

		acc = Must(vers.AddBlob(blob, "", "artifact2:v1", nil))
		MustBeSuccessful(vers.SetResource(cpi.NewResourceMeta("image2", resourcetypes.OCI_IMAGE, metav1.LocalRelation), acc, cpi.ModifyResource()))
		MustBeSuccessful(comp.AddVersion(vers))

		res = Must(vers.GetResourceByIndex(1))
		acc = Must(res.Access())
		acc = acc.GlobalAccessSpec(ctx)
		Expect(acc).NotTo(BeNil())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		o = acc.(*ociartifact.AccessSpec)
		Expect(o.ImageReference).To(Equal(TESTBASE + "/artifact2:v1@sha256:" + testhelper.DIGEST_MANIFEST))

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
		impl := Must(repocpi.GetRepositoryImplementation(repo))
		Expect(reflect.TypeOf(impl).String()).To(Equal("*genericocireg.RepositoryImpl"))

		nested := finalize.Nested()

		comp := finalizer.ClosingWith(nested, Must(repo.LookupComponent(COMPONENT)))
		vers := finalizer.ClosingWith(nested, Must(comp.NewVersion("v1")))

		m1 := compdesc.NewResourceMeta("rsc1", resourcetypes.PLAIN_TEXT, metav1.LocalRelation)
		blob := blobaccess.ForString(mime.MIME_TEXT, ocmtesthelper.S_TESTDATA)

		MustBeSuccessful(vers.SetResourceBlob(m1, blob, "", nil))
		MustBeSuccessful(comp.AddVersion(vers))

		MustBeSuccessful(nested.Finalize())

		// modify resource in component
		vers = finalizer.ClosingWith(nested, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
		blob = blobaccess.ForString(mime.MIME_TEXT, "otherdata")
		MustBeSuccessful(vers.SetResourceBlob(m1, blob, "", nil))
		MustBeSuccessful(vers.Update())
		MustBeSuccessful(nested.Finalize())

		// check content
		vers = finalizer.ClosingWith(nested, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
		r := Must(vers.GetResource(metav1.NewIdentity("rsc1")))
		data := Must(ocmutils.GetResourceData(r))
		Expect(string(data)).To(Equal("otherdata"))
		MustBeSuccessful(nested.Finalize())

		MustBeSuccessful(finalize.Finalize())

		ocirepo := Must(DefaultContext.OCIContext().RepositoryForSpec(ocispec))
		finalize.Close(ocirepo)

		art := Must(ocirepo.LookupArtifact("component-descriptors/"+COMPONENT, "v1"))
		finalize.Close(art)

		Expect(art.GetDescriptor().IsManifest()).To(BeTrue())
		m := Must(art.GetDescriptor().Manifest())
		Expect(len(m.Layers)).To(Equal(2))
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
			m := Must(art.GetDescriptor().Manifest())
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
			m := Must(art.GetDescriptor().Manifest())
			Expect(m.Config.MediaType).To(Equal(componentmapping.LegacyComponentDescriptorConfigMimeType))
			Expect(len(m.Layers)).To(Equal(1))
			Expect(m.Layers[0].MediaType).To(Equal(componentmapping.LegacyComponentDescriptorTarMimeType))
			MustBeSuccessful(finalize.Finalize())

			repo = finalizer.ClosingWith(&finalize, Must(DefaultContext.RepositoryForSpec(spec)))
			vers = finalizer.ClosingWith(&finalize, Must(repo.LookupComponentVersion(COMPONENT, "v1")))
			Expect(string(vers.GetDescriptor().Provider.Name)).To(Equal("acme.org"))
		})
	})

	Context("repo urls", func() {
		It("creates scheme based repo", func() {
			ctx := ocm.New()

			spec := ocmreg.NewRepositorySpec("http://127.0.0.1:5000/ocm")
			repo := Must(ctx.RepositoryForSpec(spec))
			defer Close(repo, "repo")

			ocirepo := genericocireg.GetOCIRepository(repo)
			Expect(ocirepo).NotTo(BeNil())
			impl := Must(ocicpi.GetRepositoryImplementation(ocirepo))

			Expect(impl).NotTo(BeNil())

			Expect(impl.(*ocireg.RepositoryImpl).GetBaseURL()).To(Equal("http://127.0.0.1:5000"))
			Expect(impl.(*ocireg.RepositoryImpl).GetRef("repo/path", "1.0.0")).To(Equal("127.0.0.1:5000/repo/path:1.0.0"))
		})
	})
})
