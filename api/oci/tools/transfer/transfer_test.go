package transfer_test

import (
	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "ocm.software/ocm/api/helper/builder"
	. "ocm.software/ocm/api/oci/testhelper"

	"github.com/mandelsoft/goutils/finalizer"

	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/ctf"
	"ocm.software/ocm/api/oci/tools/transfer"
	"ocm.software/ocm/api/oci/tools/transfer/filters"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

const (
	OUT     = "/tmp/res"
	OCIPATH = "/tmp/oci"
)

var _ = Describe("transfer OCI artifacts", func() {
	var env *Builder
	var idesc *artdesc.Descriptor

	BeforeEach(func() {
		env = NewBuilder()
		env.OCICommonTransport(OCIPATH, accessio.FormatDirectory, func() {
			idesc = OCIIndex1(env)
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("transfers index", func() {
		// index implicitly tests transfer of simple manifest, also
		var finalize finalizer.Finalizer
		defer Defer(finalize.Finalize)

		src := Must(ctf.Open(env.OCIContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
		finalize.Close(src, "source")
		art := Must(src.LookupArtifact(OCINAMESPACE3, OCIINDEXVERSION))
		finalize.Close(art, "source artifact")

		tgt := Must(ctf.Create(env.OCIContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
		defer Close(tgt, "target")
		ns := Must(tgt.LookupNamespace(OCINAMESPACE3))
		defer Close(ns, "target namespace")

		MustBeSuccessful(transfer.TransferArtifact(art, ns, OCIINDEXVERSION))

		MustBeSuccessful(finalize.Finalize())

		tart := Must(ns.GetArtifact(idesc.Digest.String()))
		defer Close(tart, "target index artifact")

		Expect(tart.IsIndex())
		manifests := tart.IndexAccess().GetDescriptor().Manifests
		Expect(len(manifests)).To(Equal(3))
		Expect(manifests[0].Digest.Encoded()).To(Equal(D_OCIMANIFEST1))
		Expect(manifests[1].Digest.Encoded()).To(Equal(D_OCIMANIFEST2))
		Expect(manifests[2].Digest.Encoded()).To(Equal(D_OCIMANIFEST2))

		nart1 := Must(ns.GetArtifact(manifests[0].Digest.String()))
		defer Close(nart1, "nested artifact 1")
		Expect(nart1.IsManifest()).To(BeTrue())
		Expect(len(nart1.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
		blob := Must(nart1.ManifestAccess().GetBlob(nart1.ManifestAccess().GetDescriptor().Layers[0].Digest))
		defer Close(blob, "layer 0 of nested artifact 1")
		data := Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER))
		Expect(manifests[0].Platform).To(Equal(&artdesc.Platform{OS: "linux", Architecture: "amd64"}))

		nart2 := Must(ns.GetArtifact(manifests[1].Digest.String()))
		defer Close(nart2, "nested artifact 2")
		Expect(nart2.IsManifest()).To(BeTrue())
		Expect(len(nart2.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
		blob = Must(nart2.ManifestAccess().GetBlob(nart2.ManifestAccess().GetDescriptor().Layers[0].Digest))
		defer Close(blob, "layer 0 of nested artifact 2")
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER2))
		Expect(manifests[1].Platform).To(Equal(&artdesc.Platform{OS: "linux", Architecture: "arm64"}))

		nart3 := Must(ns.GetArtifact(manifests[2].Digest.String()))
		defer Close(nart3, "nested artifact 3")
		Expect(nart2.IsManifest()).To(BeTrue())
		Expect(len(nart3.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
		blob = Must(nart3.ManifestAccess().GetBlob(nart2.ManifestAccess().GetDescriptor().Layers[0].Digest))
		defer Close(blob, "layer 0 of nested artifact 3")
		data = Must(blob.Get())
		Expect(string(data)).To(Equal(OCILAYER2))
		Expect(manifests[2].Platform).To(Equal(&artdesc.Platform{OS: "darwin", Architecture: "arm64"}))
	})

	Context("with filter", func() {
		It("transfers index", func() {
			// index implicitly tests transfer of simple manifest, also
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			src := Must(ctf.Open(env.OCIContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
			finalize.Close(src, "source")
			art := Must(src.LookupArtifact(OCINAMESPACE3, OCIINDEXVERSION))
			finalize.Close(art, "source artifact")

			tgt := Must(ctf.Create(env.OCIContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer Close(tgt, "target")
			ns := Must(tgt.LookupNamespace(OCINAMESPACE3))
			defer Close(ns, "target namespace")

			filter := filters.Platform("linux", "")
			MustBeSuccessful(transfer.TransferArtifactWithFilter(art, ns, filter, OCIINDEXVERSION))

			MustBeSuccessful(finalize.Finalize())

			tart := Must(ns.GetArtifact(OCIINDEXVERSION))
			defer Close(tart, "target index artifact")

			Expect(tart.IsIndex())
			manifests := tart.IndexAccess().GetDescriptor().Manifests
			Expect(len(manifests)).To(Equal(2))
			Expect(manifests[0].Digest.Encoded()).To(Equal(D_OCIMANIFEST1))
			Expect(manifests[1].Digest.Encoded()).To(Equal(D_OCIMANIFEST2))

			nart1 := Must(ns.GetArtifact(manifests[0].Digest.String()))
			defer Close(nart1, "nested artifact 1")
			Expect(nart1.IsManifest()).To(BeTrue())
			Expect(len(nart1.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
			blob := Must(nart1.ManifestAccess().GetBlob(nart1.ManifestAccess().GetDescriptor().Layers[0].Digest))
			defer Close(blob, "layer 0 of nested artifact 1")
			data := Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
			Expect(manifests[0].Platform).To(Equal(&artdesc.Platform{OS: "linux", Architecture: "amd64"}))

			nart2 := Must(ns.GetArtifact(manifests[1].Digest.String()))
			defer Close(nart2, "nested artifact 2")
			Expect(nart2.IsManifest()).To(BeTrue())
			Expect(len(nart2.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
			blob = Must(nart2.ManifestAccess().GetBlob(nart2.ManifestAccess().GetDescriptor().Layers[0].Digest))
			defer Close(blob, "layer 0 of nested artifact 2")
			data = Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER2))
			Expect(manifests[1].Platform).To(Equal(&artdesc.Platform{OS: "linux", Architecture: "arm64"}))
		})

		It("transfers index to manifest", func() {
			// index implicitly tests transfer of simple manifest, also
			var finalize finalizer.Finalizer
			defer Defer(finalize.Finalize)

			src := Must(ctf.Open(env.OCIContext(), accessobj.ACC_READONLY, OCIPATH, 0, env))
			finalize.Close(src, "source")
			art := Must(src.LookupArtifact(OCINAMESPACE3, OCIINDEXVERSION))
			finalize.Close(art, "source artifact")

			tgt := Must(ctf.Create(env.OCIContext(), accessobj.ACC_WRITABLE|accessobj.ACC_CREATE, OUT, 0o700, accessio.FormatDirectory, env))
			defer Close(tgt, "target")
			ns := Must(tgt.LookupNamespace(OCINAMESPACE3))
			defer Close(ns, "target namespace")

			filter := filters.Platform("linux", "amd64")
			MustBeSuccessful(transfer.TransferArtifactWithFilter(art, ns, filter, OCIINDEXVERSION))

			MustBeSuccessful(finalize.Finalize())

			tart := Must(ns.GetArtifact(OCIINDEXVERSION))
			defer Close(tart, "target index artifact")

			Expect(tart.IsManifest())

			nart1 := tart
			Expect(nart1.IsManifest()).To(BeTrue())
			Expect(len(nart1.ManifestAccess().GetDescriptor().Layers)).To(Equal(1))
			blob := Must(nart1.ManifestAccess().GetBlob(nart1.ManifestAccess().GetDescriptor().Layers[0].Digest))
			defer Close(blob, "layer 0 of nested artifact 1")
			data := Must(blob.Get())
			Expect(string(data)).To(Equal(OCILAYER))
		})
	})
})
