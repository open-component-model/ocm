package ocirepo_test

import (
	"encoding/json"
	"fmt"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	ocictf "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	. "ocm.software/ocm/api/oci/testhelper"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/keepblobattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/mapocirepoattr"
	"ocm.software/ocm/api/ocm/extensions/attrs/preferrelativeattr"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo/config"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/extensions/repositories/genericocireg"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

const (
	ARCH       = "/tmp/ctf"
	ARCH2      = "/tmp/ctf2"
	PROVIDER   = "mandelsoft"
	VERSION    = "v1"
	COMPONENT  = "github.com/mandelsoft/test"
	COMPONENT2 = "github.com/mandelsoft/test2"
	OUT        = "/tmp/res"
	OCIPATH    = "/tmp/oci"
	OCIHOST    = "alias"
)

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return "baseurl.io"
}

var _ = Describe("oci artifact transfer", func() {
	var env *Builder
	var ldesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder()

		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			ldesc = OCIManifest1(env)
		})

		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
						env.BlobStringData(mime.MIME_TEXT, "testdata")
					})
					env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})

		_ = ldesc
	})

	AfterEach(func() {
		env.Cleanup()
	})

	Context("with tag", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
						env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
							)
						})
					})
				})
			})
		})

		It("it should copy a resource by value and export the OCI image using a relative OCI access method", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
			preferrelativeattr.Set(env.OCMContext(), true)

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(data).To(YAMLEqual(`{"reference":"` + OCINAMESPACE + ":" + OCIVERSION + `@sha256:` + D_OCIMANIFEST1 + `","type":"relativeOciReference"}`))
			ocirepo := genericocireg.GetOCIRepository(tgt)
			Expect(ocirepo).NotTo(BeNil())

			res := Must(comp.GetResourceByIndex(1))
			s := Must(ocmutils.GetOCIArtifactRef(comp.GetContext(), res))
			Expect(s).To(Equal("ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))
			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))

			b := Must(res.BlobAccess())
			defer Close(b, "blob")

			set := Must(artifactset.OpenFromBlob(accessobj.ACC_READONLY, b, env))
			defer Close(set, "artifact")
			Expect(set.GetAnnotation(artifactset.MAINARTIFACT_ANNOTATION)).To(Equal("sha256:" + D_OCIMANIFEST1))
		})

		DescribeTable("it should copy a resource by value and export the OCI image using a relative OCI access method", func(prefer bool, repos []string, relative bool) {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			o := config.New()
			o.UploadOptions = config.UploadOptions{PreferRelativeAccess: prefer, Repositories: repos}
			env.ConfigContext().ApplyConfig(o, "manual")
			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			if relative {
				Expect(data).To(YAMLEqual(`{"reference":"` + OCINAMESPACE + ":" + OCIVERSION + `@sha256:` + D_OCIMANIFEST1 + `","type":"relativeOciReference"}`))

				ocirepo := genericocireg.GetOCIRepository(tgt)
				Expect(ocirepo).NotTo(BeNil())

				res := Must(comp.GetResourceByIndex(1))
				s := Must(ocmutils.GetOCIArtifactRef(comp.GetContext(), res))
				if relative {
					// cannot be faked
					Expect(s).To(Equal("ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))
				} else {
					Expect(s).To(Equal("baseurl.io/ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))
				}
				art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
				defer Close(art, "artifact")

				man := MustBeNonNil(art.ManifestAccess())
				Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
				Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

				blob := Must(man.GetBlob(ldesc.Digest))
				data = Must(blob.Get())
				Expect(string(data)).To(Equal(OCILAYER))

				b := Must(res.BlobAccess())
				defer Close(b, "blob")

				set := Must(artifactset.OpenFromBlob(accessobj.ACC_READONLY, b, env))
				defer Close(set, "artifact")
				Expect(set.GetAnnotation(artifactset.MAINARTIFACT_ANNOTATION)).To(Equal("sha256:" + D_OCIMANIFEST1))
			} else {
				// cannot access faked local repository URL baseurl.io
				Expect(data).To(YAMLEqual(`{"imageReference":"baseurl.io/` + OCINAMESPACE + ":" + OCIVERSION + `@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))
			}
		},
			Entry("none", false, nil, false),
			Entry("baseurl.io", true, []string{"baseurl.io"}, true),
			Entry("baseurl.de", true, []string{"baseurl.de"}, false),
			Entry("all", true, nil, true),
		)

		It("it should copy a resource by value and export the OCI image but keep the local blob", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
			keepblobattr.Set(env.OCMContext(), true)

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"globalAccess":{"imageReference":"baseurl.io/ocm/value:v2.0@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"},"localReference":"sha256:` + H_OCIARCHMANIFEST1 + `","mediaType":"application/vnd.oci.image.manifest.v1+tar+gzip","referenceName":"ocm/value:v2.0","type":"localBlob"}`))
			ocirepo := genericocireg.GetOCIRepository(tgt)
			Expect(ocirepo).NotTo(BeNil())

			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})

		It("it should copy a resource by value and export the OCI image", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/ocm/value:v2.0@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			ocirepo := genericocireg.GetOCIRepository(tgt)
			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})

		It("it should copy a resource by value and export the OCI image with hashed repo name", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)
			mapocirepoattr.Set(env.OCMContext(), &mapocirepoattr.Attribute{Mode: mapocirepoattr.ShortHashMode, Always: true})
			rdigest := "e9b6af2174cb2fb78b2882a1f487b01295b8f6bfa7e4c1ceb350440104c9ce65"

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/` + rdigest[:8] + `/value:v2.0@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			namespace := rdigest[:8] + "/value"
			ocirepo := genericocireg.GetOCIRepository(tgt)
			art := Must(ocirepo.LookupArtifact(namespace, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})

		It("it should copy a resource by value and export the OCI image with hashed repo name and prefix", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)
			prefix := "ocm"
			mapocirepoattr.Set(env.OCMContext(), &mapocirepoattr.Attribute{Mode: mapocirepoattr.ShortHashMode, Always: true, Prefix: &prefix})
			rdigest := "e9b6af2174cb2fb78b2882a1f487b01295b8f6bfa7e4c1ceb350440104c9ce65"

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/ocm/` + rdigest[:8] + `/value:v2.0@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			namespace := "ocm/" + rdigest[:8] + "/value"
			ocirepo := genericocireg.GetOCIRepository(tgt)
			art := Must(ocirepo.LookupArtifact(namespace, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})
	})

	Context("with digest", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
						env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1)),
							)
						})
					})
				})
			})
		})

		It("it should copy a resource by value and export the OCI image", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/ocm/value@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			ocirepo := genericocireg.GetOCIRepository(tgt)
			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})
	})

	Context("with tag + digest", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
						env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION+"@sha256:"+D_OCIMANIFEST1)),
							)
						})
					})
				})
			})
		})

		It("it should copy a resource by value and export the OCI image", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/ocm/value:v2.0@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			ocirepo := genericocireg.GetOCIRepository(tgt)

			Expect(Must(ocirepo.ExistsArtifact(OCINAMESPACE, OCIVERSION))).To(BeTrue())
			Expect(Must(ocirepo.ExistsArtifact(OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1))).To(BeTrue())

			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, OCIVERSION))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})
	})

	Context("with wrong tag + digest", func() {
		BeforeEach(func() {
			env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
				env.Component(COMPONENT, func() {
					env.Version(VERSION, func() {
						env.Provider(PROVIDER)
						env.Resource("testdata", "", "PlainText", metav1.LocalRelation, func() {
							env.BlobStringData(mime.MIME_TEXT, "testdata")
						})
						env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
							env.Access(
								ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, "dummy"+"@sha256:"+D_OCIMANIFEST1)),
							)
						})
					})
				})
			})
		})

		It("it should copy a resource by value and export the OCI image", func() {
			env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
				cpi.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), cpi.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))

			src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
			cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
			tgt := Must(ctf.Create(env.OCMContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer tgt.Close()

			opts := &standard.Options{}
			opts.SetResourcesByValue(true)
			handler := standard.NewDefaultHandler(opts)

			MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, tgt, handler))
			Expect(env.DirExists(OUT)).To(BeTrue())

			list := Must(tgt.ComponentLister().GetComponents("", true))
			Expect(list).To(Equal([]string{COMPONENT}))
			comp := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
			Expect(len(comp.GetDescriptor().Resources)).To(Equal(2))
			data := Must(json.Marshal(comp.GetDescriptor().Resources[1].Access))

			fmt.Printf("%s\n", string(data))
			// provide the fake name tag in target repo
			Expect(string(data)).To(StringEqualWithContext(`{"imageReference":"baseurl.io/ocm/value:dummy@sha256:` + D_OCIMANIFEST1 + `","type":"ociArtifact"}`))

			ocirepo := genericocireg.GetOCIRepository(tgt)

			Expect(Must(ocirepo.ExistsArtifact(OCINAMESPACE, "dummy"))).To(BeTrue())
			Expect(Must(ocirepo.ExistsArtifact(OCINAMESPACE, "@sha256:"+D_OCIMANIFEST1))).To(BeTrue())

			art := Must(ocirepo.LookupArtifact(OCINAMESPACE, "dummy"))
			defer Close(art, "artifact")

			man := MustBeNonNil(art.ManifestAccess())
			Expect(len(man.GetDescriptor().Layers)).To(Equal(1))
			Expect(man.GetDescriptor().Layers[0].Digest).To(Equal(ldesc.Digest))

			blob := Must(man.GetBlob(ldesc.Digest))
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})
	})
})
