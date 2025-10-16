package ocirepo_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	ctfoci "ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/oci/grammar"
	. "ocm.software/ocm/api/oci/testhelper"
	v1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/attrs/ociuploadattr"
	"ocm.software/ocm/api/ocm/extensions/blobhandler"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	ctfocm "ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	COMP   = "github.com/compa"
	VERS   = "1.0.0"
	CA     = "ca"
	CTF    = "ctf"
	COPY   = "ctf.copy"
	TARGET = "/tmp/target"
)

const (
	OCIHOST = "alias"
	OCIPATH = "/tmp/source"
)

var _ = Describe("upload", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()

		// fake OCI registry
		FakeOCIRepo(env, OCIPATH, OCIHOST)

		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			OCIManifest1(env)
		})

		env.OCICommonTransport(TARGET, accessio.FormatDirectory)

		env.ComponentArchive(CA, accessio.FormatDirectory, COMP, VERS, func() {
			env.Provider("mandelsoft")
			env.Resource("value", "", resourcetypes.OCI_IMAGE, v1.LocalRelation, func() {
				env.Access(
					ociartifact.New(oci.StandardOCIRef(OCIHOST+".alias", OCINAMESPACE, OCIVERSION)),
				)
			})
		})

		ca := Must(comparch.Open(env.OCMContext(), accessobj.ACC_READONLY, CA, 0, env))
		oca := accessio.OnceCloser(ca)
		defer Close(oca)

		ctf := Must(ctfocm.Create(env.OCMContext(), accessobj.ACC_CREATE, CTF, 0o700, env))
		octf := accessio.OnceCloser(ctf)
		defer Close(octf)

		handler := Must(standard.New(standard.ResourcesByValue()))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, ca, ctf, handler))

		// now we have a transport archive with local blob for the image
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("validated original oci manifest", func() {
		ctx := env.OCMContext()

		ocirepo := Must(ctfoci.Open(ctx, accessobj.ACC_READONLY, OCIPATH, 0o700, env))
		defer Close(ocirepo, "ocoirepo")

		ns := Must(ocirepo.LookupNamespace(OCINAMESPACE))
		defer Close(ns, "namespace")

		art := Must(ns.GetArtifact(OCIVERSION))
		defer Close(art, "artifact")

		Expect(art.Digest().Encoded()).To(Equal(D_OCIMANIFEST1))
	})

	It("validated original digest", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0o700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		defer Close(cv, "component version")

		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		Expect(ra.Meta().Digest).To(Equal(DS_OCIMANIFEST1))
	})

	It("transfers oci artifact", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0o700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy := Must(ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0o700, env))
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		ociuploadattr.Set(ctx, attr)

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, copy, nil))

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra = Must(cv2.GetResourceByIndex(0))
		Expect(ra.Meta().Digest).To(Equal(DS_OCIMANIFEST1))
		acc = Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartifact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))

		attr.Close()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtifact("copy/ocm/value", "v2.0")).To(BeTrue())
	})

	It("transfers oci artifact with named handler and object config", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0o700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy := Must(ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0o700, env))
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		MustBeSuccessful(blobhandler.RegisterHandlerByName(ctx, "ocm/ociArtifacts", attr))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, copy, nil))

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra = Must(cv2.GetResourceByIndex(0))
		acc = Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartifact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))

		// attr.Close()
		env.OCMContext().Finalize()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtifact("copy/ocm/value", "v2.0")).To(BeTrue())
	})

	It("transfers oci artifact with named handler and string config", func() {
		ctx := env.OCMContext()

		ctf := Must(ctfocm.Open(ctx, accessobj.ACC_READONLY, CTF, 0o700, env))
		defer Close(ctf, "ctf")

		cv := Must(ctf.LookupComponentVersion(COMP, VERS))
		ocv := accessio.OnceCloser(cv)
		defer Close(ocv)
		ra := Must(cv.GetResourceByIndex(0))
		acc := Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(localblob.Type))

		// transfer component
		copy := Must(ctfocm.Create(ctx, accessobj.ACC_CREATE, COPY, 0o700, env))
		ocopy := accessio.OnceCloser(copy)
		defer Close(ocopy)

		// prepare upload to target OCI repo
		// attr := ociuploadattr.New(TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy")
		attr := TARGET + grammar.RepositorySeparator + grammar.RepositorySeparator + "copy"
		MustBeSuccessful(blobhandler.RegisterHandlerByName(ctx, "ocm/ociArtifacts", attr))

		MustBeSuccessful(transfer.TransferVersion(nil, nil, cv, copy, nil))

		// check type
		cv2 := Must(copy.LookupComponentVersion(COMP, VERS))
		ocv2 := accessio.OnceCloser(cv2)
		defer Close(ocv2)
		ra = Must(cv2.GetResourceByIndex(0))
		acc = Must(ra.Access())
		Expect(acc.GetKind()).To(Equal(ociartifact.Type))
		val := Must(ctx.AccessSpecForSpec(acc))
		// TODO: the result is invalid for ctf: better handling for ctf refs
		Expect(val.(*ociartifact.AccessSpec).ImageReference).To(Equal("/tmp/target//copy/ocm/value:v2.0@sha256:" + D_OCIMANIFEST1))

		// attr.Close()
		env.OCMContext().Finalize()
		target, err := ctfoci.Open(ctx.OCIContext(), accessobj.ACC_READONLY, TARGET, 0, env)
		Expect(err).To(Succeed())
		defer Close(target)
		Expect(target.ExistsArtifact("copy/ocm/value", "v2.0")).To(BeTrue())
	})
})
