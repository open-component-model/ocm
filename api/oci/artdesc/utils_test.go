package artdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"ocm.software/ocm/api/oci/artdesc"
)

var _ = Describe("utils", func() {
	It("strips media type", func() {
		Expect(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest)).To(Equal("application/vnd.oci.image.manifest.v1"))
		Expect(artdesc.ToContentMediaType(artdesc.MediaTypeImageIndex)).To(Equal("application/vnd.oci.image.index.v1"))
	})

	It("maps to descriptor media typ", func() {
		Expect(artdesc.ToDescriptorMediaType(artdesc.ToContentMediaType(artdesc.MediaTypeImageManifest) + "+tar+gzip")).To(Equal(artdesc.MediaTypeImageManifest))
		Expect(artdesc.ToDescriptorMediaType(artdesc.ToContentMediaType(artdesc.MediaTypeImageIndex) + "+tar+gzip")).To(Equal(artdesc.MediaTypeImageIndex))
	})
})
