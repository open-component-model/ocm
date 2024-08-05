package testhelper

import (
	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/finalizer"
	"github.com/mandelsoft/goutils/testutils"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/cpi"
	"ocm.software/ocm/api/utils/blobaccess"
	"ocm.software/ocm/api/utils/mime"
)

const (
	TAG             = "v1"
	DIGEST_MANIFEST = "3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a"
	DIGEST_LAYER    = "810ff2fb242a5dee4220f2cb0e6a519891fb67f2f828a6cab4ef8894633b1f50"
	DIGEST_CONFIG   = "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
)

func DefaultManifestFill(n cpi.NamespaceAccess) {
	var finalize finalizer.Finalizer
	defer testutils.Defer(finalize.Finalize)

	art := NewArtifact(n, &finalize)
	blob := testutils.Must(n.AddArtifact(art))
	n.AddTags(blob.Digest(), TAG)
}

func NewArtifact(n cpi.NamespaceAccess, finalize *finalizer.Finalizer) cpi.ArtifactAccess {
	art := testutils.Must(n.NewArtifact())
	finalize.Close(art)
	Expect(art.AddLayer(blobaccess.ForString(mime.MIME_OCTET, "testdata"), nil)).To(Equal(0))
	desc := testutils.Must(art.Manifest())
	Expect(desc).NotTo(BeNil())

	Expect(desc.Layers[0].Digest).To(Equal(digest.FromString("testdata")))
	Expect(desc.Layers[0].MediaType).To(Equal(mime.MIME_OCTET))
	Expect(desc.Layers[0].Size).To(Equal(int64(8)))

	config := blobaccess.ForData(mime.MIME_OCTET, []byte("{}"))
	testutils.MustBeSuccessful(n.AddBlob(config))
	desc.Config = *artdesc.DefaultBlobDescriptor(config)
	return art
}

func CheckArtifact(art oci.ArtifactAccess) {
	Expect(art.IsManifest()).To(BeTrue())
	blob := testutils.Must(art.GetBlob("sha256:" + DIGEST_LAYER))
	Expect(blob.Get()).To(Equal([]byte("testdata")))
	Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
	blob = testutils.Must(art.GetBlob("sha256:" + DIGEST_CONFIG))
	Expect(blob.Get()).To(Equal([]byte("{}")))
	Expect(blob.MimeType()).To(Equal(mime.MIME_OCTET))
}
