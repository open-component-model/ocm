package signing_test

import (
	"strings"

	. "github.com/mandelsoft/goutils/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	. "ocm.software/ocm/api/helper/builder"
	"ocm.software/ocm/api/oci/artdesc"
	"ocm.software/ocm/api/oci/extensions/repositories/artifactset"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	resourcetypes "ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/extensions/digester/digesters/artifact"
	"ocm.software/ocm/api/ocm/extensions/repositories/ctf"
	"ocm.software/ocm/api/ocm/tools/signing"
	"ocm.software/ocm/api/tech/signing/hasher/sha256"
	"ocm.software/ocm/api/utils/accessio"
	"ocm.software/ocm/api/utils/accessobj"
)

var _ = Describe("Digest Test Environment", func() {
	var env *Builder

	BeforeEach(func() {
		env = NewBuilder()
		env.OCMCommonTransport(ARCH, accessio.FormatDirectory, func() {
			env.ComponentVersion(COMPONENTA, VERSION, func() {
				env.Resource("image", VERSION, resourcetypes.OCI_ARTIFACT, metav1.LocalRelation, func() {
					env.ArtifactSetBlob(VERSION, func() {
						env.Manifest(VERSION, func() {
							env.Config(func() {
								env.BlobStringData(ociv1.MediaTypeImageConfig, "{}")
							})
							env.Layer(func() {
								env.BlobStringData(ociv1.MediaTypeImageLayerGzip, "fake")
							})
						})
					})
				})
			})
		})
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("created oci artifact", func() {
		repo := Must(ctf.Open(env.OCMContext(), accessobj.ACC_READONLY, ARCH, 0, env))
		defer Close(repo, "repo")

		cv := Must(repo.LookupComponentVersion(COMPONENTA, VERSION))
		defer Close(cv, "cv")

		Expect(len(cv.GetDescriptor().Resources)).To(Equal(1))

		rsc := Must(cv.GetResourceByIndex(0))

		m := Must(rsc.AccessMethod())
		defer Close(m, "meth")
		Expect(m.MimeType()).To(Equal(artifactset.MediaType(artdesc.MediaTypeImageManifest)))

		dig := rsc.Meta().Digest
		Expect(dig).NotTo(BeNil())
		Expect(dig.HashAlgorithm).To(Equal(sha256.Algorithm))
		Expect(dig.NormalisationAlgorithm).To(Equal(artifact.OciArtifactDigestV1))

		Expect(Must(signing.VerifyResourceDigest(cv, 0, m))).To(BeTrue())

		orig := dig.Value
		dig.Value = strings.Replace(dig.Value, "a", "b", -1)
		done, err := signing.VerifyResourceDigest(cv, 0, m)
		dig.Value = orig
		Expect(done).To(BeTrue())
		Expect(err).To(HaveOccurred())
	})
})
