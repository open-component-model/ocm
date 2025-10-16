package oci_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/opencontainers/go-digest"

	"ocm.software/ocm/api/oci"
)

func CheckArt(ref string, exp *oci.ArtSpec) {
	spec, err := oci.ParseArt(ref)
	if exp == nil {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).To(Succeed())
		Expect(spec).To(Equal(exp))
	}
}

var _ = Describe("art parsing", func() {
	digest := digest.Digest("sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	tag := "v1"

	It("succeeds", func() {
		CheckArt("ubuntu", &oci.ArtSpec{Repository: "ubuntu"})
		CheckArt("ubuntu/test", &oci.ArtSpec{Repository: "ubuntu/test"})
		CheckArt("ubuntu/test@"+digest.String(), &oci.ArtSpec{Repository: "ubuntu/test", ArtVersion: oci.ArtVersion{Digest: &digest}})
		CheckArt("ubuntu/test:"+tag, &oci.ArtSpec{Repository: "ubuntu/test", ArtVersion: oci.ArtVersion{Tag: &tag}})
		CheckArt("ubuntu/test:"+tag+"@"+digest.String(), &oci.ArtSpec{Repository: "ubuntu/test", ArtVersion: oci.ArtVersion{Digest: &digest, Tag: &tag}})
	})

	It("fails", func() {
		CheckArt("ubu@ntu", nil)
		CheckArt("ubu@sha256:123", nil)
	})
})
