package testhelper

import (
	"bytes"
	"io"

	. "github.com/onsi/gomega"

	"github.com/mandelsoft/goutils/testutils"
	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/digester/digesters/artifact"
	"ocm.software/ocm/api/ocm/ocmutils"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
	"ocm.software/ocm/api/utils/mime"
)

// CheckOCIArtifactBlob validates that acc provides a synthesized OCI artifact
// set archive whose decoded content matches the expected fixture: expectedRoot
// is the hex (without "sha256:") of the main manifest, expectedLayer the single
// text layer's content. It verifies the archive integrity independently of the
// compressor used to produce the "+tar+gzip" blob, and returns the actual
// "sha256:"-prefixed digest of the stored archive bytes so callers can assert
// the local reference without pinning a compressor-specific value.
func CheckOCIArtifactBlob(acc ocm.AccessProvider, expectedRoot, expectedLayer string) string {
	data, err := ocmutils.GetResourceData(acc)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	archiveDigest := digest.FromBytes(data)

	set, err := artifactset.Open(accessobj.ACC_READONLY, "", 0, accessio.Reader(io.NopCloser(bytes.NewReader(data))))
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	defer set.Close()

	main := set.GetMain()
	ExpectWithOffset(1, main.String()).To(Equal("sha256:" + expectedRoot))

	// independently re-hash the main manifest blob instead of trusting the index
	_, manifestData, err := set.GetBlobData(main)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	defer manifestData.Close()
	manifestBytes, err := manifestData.Get()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, digest.FromBytes(manifestBytes).String()).To(Equal("sha256:" + expectedRoot))

	// config blob is the fixed "{}" document
	_, configData, err := set.GetBlobData(digest.FromBytes([]byte("{}")))
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	defer configData.Close()
	configBytes, err := configData.Get()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, configBytes).To(Equal([]byte("{}")))

	// layer blob is the fixed text content
	_, layerData, err := set.GetBlobData(digest.FromBytes([]byte(expectedLayer)))
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	defer layerData.Close()
	layerBytes, err := layerData.Get()
	ExpectWithOffset(1, err).ToNot(HaveOccurred())
	ExpectWithOffset(1, layerBytes).To(Equal([]byte(expectedLayer)))

	return archiveDigest.String()
}

////////////////////////////////////////////////////////////////////////////////
// manifestlaver

const (
	OCINAMESPACE = "ocm/value"
	OCIVERSION   = "v2.0"
	OCILAYER     = "manifestlayer"
)

var OCIDigests = testutils.Substitutions{
	"D_OCIMANIFEST1": D_OCIMANIFEST1,
	"D_OCIMANIFEST2": D_OCIMANIFEST2,
}

func OCIManifest1(env *builder.Builder) *artdesc.Descriptor {
	_, ldesc := OCIManifest1For(env, OCINAMESPACE, OCIVERSION)
	return ldesc
}

func OCIManifest1For(env *builder.Builder, ns, tag string, nested ...func()) (*artdesc.Descriptor, *artdesc.Descriptor) {
	var ldesc *artdesc.Descriptor
	var mdesc *artdesc.Descriptor

	env.Namespace(ns, func() {
		mdesc = env.Manifest(tag, append(nested, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, OCILAYER)
			})
		})...)
	})
	return mdesc, ldesc
}

func OCIArtifactResource1(env *builder.Builder, name string, host string, funcs ...func()) {
	env.Resource(name, "", resourcetypes.OCI_IMAGE, metav1.LocalRelation, func() {
		env.Access(
			ociartifact.New(oci.StandardOCIRef(host+".alias", OCINAMESPACE, OCIVERSION)),
		)
		env.Configure(funcs...)
	})
}

const (
	D_OCIMANIFEST1 = "0c4abdb72cf59cb4b77f4aacb4775f9f546ebc3face189b2224a966c8826ca9f"
)

var DS_OCIMANIFEST1 = &metav1.DigestSpec{
	HashAlgorithm:          sha256.Algorithm,
	NormalisationAlgorithm: artifact.OciArtifactDigestV1,
	Value:                  D_OCIMANIFEST1,
}

////////////////////////////////////////////////////////////////////////////////
// otherlayer

const (
	OCINAMESPACE2 = "ocm/ref"
	OCILAYER2     = "otherlayer"
)

var DS_OCIMANIFEST2 = &metav1.DigestSpec{
	HashAlgorithm:          sha256.Algorithm,
	NormalisationAlgorithm: artifact.OciArtifactDigestV1,
	Value:                  D_OCIMANIFEST2,
}

func OCIManifest2(env *builder.Builder) *artdesc.Descriptor {
	_, ldesc := OCIManifest2For(env, OCINAMESPACE2, OCIVERSION)
	return ldesc
}

func OCIManifest2For(env *builder.Builder, ns, tag string, nested ...func()) (*artdesc.Descriptor, *artdesc.Descriptor) {
	var ldesc *artdesc.Descriptor
	var mdesc *artdesc.Descriptor

	env.Namespace(ns, func() {
		mdesc = env.Manifest(tag, append(nested, func() {
			env.Config(func() {
				env.BlobStringData(mime.MIME_JSON, "{}")
			})
			ldesc = env.Layer(func() {
				env.BlobStringData(mime.MIME_TEXT, OCILAYER2)
			})
		})...)
	})
	return mdesc, ldesc
}

const (
	D_OCIMANIFEST2 = "c2d2dca275c33c1270dea6168a002d67c0e98780d7a54960758139ae19984bd7"
)

////////////////////////////////////////////////////////////////////////////////

const (
	OCIINDEXVERSION = "v2.0-index"
	OCINAMESPACE3   = "ocm/index"
)

func OCIIndex1(env *builder.Builder) *artdesc.Descriptor {
	var idesc *artdesc.Descriptor

	a1, _ := OCIManifest1For(env, OCINAMESPACE3, "", func() { env.Platform("linux", "amd64") })
	a2, _ := OCIManifest2For(env, OCINAMESPACE3, "", func() { env.Platform("linux", "arm64") })
	a3, _ := OCIManifest2For(env, OCINAMESPACE3, "", func() { env.Platform("darwin", "arm64") })

	env.Namespace(OCINAMESPACE3, func() {
		idesc = env.Index(OCIINDEXVERSION, func() {
			env.Artifact(a1)
			env.Artifact(a2)
			env.Artifact(a3)
		})
	})
	return idesc
}
