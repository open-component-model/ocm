package standard_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	ocictf "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	. "ocm.software/ocm/api/oci/testhelper"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/relativeociref"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/keepblobattr"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	storagecontext "ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci"
	"ocm.software/ocm/api/ocm/extensions/blobhandler/handlers/oci/ocirepo"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	common "ocm.software/ocm/api/utils/misc"
)

const OCIHOST2 = "target"

var _ = Describe("value transport with relative ocireg", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		env.RSAKeyPair(SIGNATURE)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		FakeOCIRepo(env, OCIPATH, OCIHOST)
		FakeOCIRepo(env, ARCH, OCIHOST2)

		env.OCMCommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			env.Component(COMPONENT, func() {
				env.Version(VERSION, func() {
					env.Provider(PROVIDER)
					env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
						env.Access(
							relativeociref.New(oci.RelativeOCIRef(OCINAMESPACE, OCIVERSION)),
						)
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("it should use additional resolver to resolve component ref", func() {
		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
		defer Close(src, "src")

		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "cv")

		r := Must(cv.GetResourceByIndex(0))
		CheckBlob(r, D_OCIMANIFEST1, 628)
	})

	DescribeTable("transfers per value", func(keep bool, mod func(env *Builder), dig string, size int, opts ...transferhandler.TransferOption) {
		env.OCMContext().BlobHandlers().Register(ocirepo.NewArtifactHandler(FakeOCIRegBaseFunction),
			blobhandler.ForRepo(oci.CONTEXT_TYPE, ocictf.Type), blobhandler.ForMimeType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)))
		keepblobattr.Set(env.OCMContext(), keep)

		p, buf := common.NewBufferedPrinter()
		topts := append([]transferhandler.TransferOption{
			standard.ResourcesByValue(), transfer.WithPrinter(p),
		}, opts...)

		src := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
		defer Close(src, "src")

		cv := Must(src.LookupComponentVersion(COMPONENT, VERSION))
		defer Close(cv, "cv")

		mod(env)

		tgt := Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		ctgt := accessio.OnceCloser(tgt)
		defer Close(ctgt, "tgt")

		MustBeSuccessful(transfer.Transfer(cv, tgt, topts...))

		options := &standard.Options{}
		transferhandler.ApplyOptions(options, topts...)

		out := `
  transferring version "github.com/mandelsoft/test:v1"...
  ...resource 0 artifact\[ociImage\]\(ocm/value:v2.0\)...
  ...adding component version...
`
		if options.IsOverwrite() {
			out = `
  transferring version "github.com/mandelsoft/test:v1"...
  warning:   version "github.com/mandelsoft/test:v1" already present, but differs because some artifact.*changed \(transport enforced by overwrite option\)
  ...resource 0 artifact\[ociImage\]\(ocm/value:v2.0\).*
  ...adding component version...
`
		}
		Expect(string(buf.Bytes())).To(StringMatchTrimmedWithContext(utils.Crop(out, 2)))
		MustBeSuccessful(ctgt.Close())

		tgt = Must(ctf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
		ctgt = accessio.OnceCloser(tgt)
		defer Close(ctgt, "tgt2")

		tcv := Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		ctcv := accessio.OnceCloser(tcv)
		defer Close(ctcv, "tcv")

		r := Must(tcv.GetResourceByIndex(0))
		acc := Must(r.Access())

		atype := ociartifact.Type
		if keep {
			atype = localblob.Type
		}
		Expect(acc.GetKind()).To(Equal(atype))

		info := acc.Info(env.OCMContext())
		if keep {
			Expect(info.Info).To(Equal("sha256:" + H_OCIARCHMANIFEST1))

		} else {
			Expect(info.Host).To(Equal(OCIHOST2 + ".alias"))
			Expect(info.Info).To(Equal("ocm/value:v2.0"))
		}

		CheckBlob(r, dig, size)

		MustBeSuccessful(ctcv.Close())
		MustBeSuccessful(ctgt.Close())

		CheckAritifact(env, dig, size)

		// re-transport
		buf.Reset()
		tgt = Must(ctf.Open(env, accessobj.ACC_WRITABLE, ARCH, 0, env))
		ctgt = accessio.OnceCloser(tgt)
		defer Close(ctgt, "tgt3")

		MustBeSuccessful(transfer.Transfer(cv, tgt, topts...))

		tcv = Must(tgt.LookupComponentVersion(COMPONENT, VERSION))
		ctcv = accessio.OnceCloser(tcv)
		defer Close(ctcv, "ctcv")

		r = Must(tcv.GetResourceByIndex(0))
		acc = Must(r.Access())
		Expect(acc.GetKind()).To(Equal(atype))

		info = acc.Info(env.OCMContext())
		if keep {
			Expect(info.Info).To(Equal("sha256:" + H_OCIARCHMANIFEST1))

		} else {
			Expect(info.Info).To(Equal("ocm/value:v2.0"))
		}

		MustBeSuccessful(ctcv.Close())
		MustBeSuccessful(ctgt.Close())

		CheckAritifact(env, dig, size)

	},
		Entry("empty target", false, EmptyTarget, D_OCIMANIFEST1, 628),
		Entry("identical target", false, IdenticalTarget, D_OCIMANIFEST1, 628),
		Entry("different target", false, DifferentTarget, D_OCIMANIFEST1, 628),
		Entry("different CV", false, DifferentCV, D_OCIMANIFEST1, 628, standard.Overwrite()),
		Entry("different namespace", false, DifferentNamespace, D_OCIMANIFEST1, 628, standard.Overwrite()),
		Entry("different name", false, DifferentName, D_OCIMANIFEST1, 628, standard.Overwrite()),
		Entry("keep, empty target", true, EmptyTarget, D_OCIMANIFEST1, 628),

		Entry("keep, identical target", true, IdenticalTarget, D_OCIMANIFEST1, 628),
		Entry("keep, different target", true, DifferentTarget, D_OCIMANIFEST1, 628),
		Entry("keep, different CV", true, DifferentCV, D_OCIMANIFEST1, 628, standard.Overwrite()),
		Entry("keep, different namespace", true, DifferentNamespace, D_OCIMANIFEST1, 628, standard.Overwrite()),
		Entry("keep, different name", true, DifferentName, D_OCIMANIFEST1, 628, standard.Overwrite()),
	)
})

func EmptyTarget(env *Builder) {
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory)
}

func IdenticalTarget(env *Builder) {
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
		OCIManifest1For(env, OCINAMESPACE, OCIVERSION)
	})
}

func DifferentTarget(env *Builder) {
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
		OCIManifest2For(env, OCINAMESPACE, OCIVERSION)
	})
}

func DifferentCV(env *Builder) {
	env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
		OCIManifest2For(env, OCINAMESPACE, OCIVERSION)
	})
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
		env.Component(COMPONENT, func() {
			env.Version(VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
					env.Access(
						ociartifact.New(oci.StandardOCIRef(OCIHOST2+".alias", OCINAMESPACE, OCIVERSION)),
					)
				})
			})
		})
	})
}

func DifferentNamespace(env *Builder) {
	env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
		OCIManifest2For(env, OCINAMESPACE2, OCIVERSION)
	})
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
		env.Component(COMPONENT, func() {
			env.Version(VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("artifact", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
					env.Access(
						relativeociref.New(oci.RelativeOCIRef(OCINAMESPACE2, OCIVERSION)),
					)
				})
			})
		})
	})
}

func DifferentName(env *Builder) {
	env.OCICommonTransport(ARCH, accessio.FormatDirectory, func() {
		OCIManifest1For(env, OCINAMESPACE, OCIVERSION)
	})
	env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
		env.Component(COMPONENT, func() {
			env.Version(VERSION, func() {
				env.Provider(PROVIDER)
				env.Resource("other", "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
					env.Access(
						ociartifact.New(oci.StandardOCIRef(OCIHOST2+".alias", OCINAMESPACE, OCIVERSION)),
					)
				})
			})
		})
	})
}

func FakeOCIRegBaseFunction(ctx *storagecontext.StorageContext) string {
	return OCIHOST2 + ".alias"
}

func CheckBlob(r ocm.ResourceAccess, dig string, size int) {
	blob := Must(r.BlobAccess())
	defer Close(blob, "blob")

	ExpectWithOffset(1, int(blob.Size())).To(Equal(size))
	set := Must(artifactset.OpenFromBlob(accessobj.ACC_READONLY, blob))
	defer Close(set, "set")

	digest := set.GetMain()
	ExpectWithOffset(1, digest.Hex()).To(Equal(dig))

	acc := Must(set.GetArtifact(digest.String()))
	defer Close(acc, "acc")

	ExpectWithOffset(1, acc.IsManifest()).To(BeTrue())
	ExpectWithOffset(1, acc.Digest().Hex()).To(Equal(dig))
}

func CheckAritifact(env *Builder, dig string, size int) {
	repo := Must(ocictf.Open(env, accessobj.ACC_READONLY, ARCH, 0, env))
	defer Close(repo, "oci repo")

	art := Must(repo.LookupArtifact(OCINAMESPACE, OCIVERSION))
	defer Close(art, "art")

	ExpectWithOffset(1, art.IsManifest()).To(BeTrue())
	ExpectWithOffset(1, art.Digest().Hex()).To(Equal(dig))
}
