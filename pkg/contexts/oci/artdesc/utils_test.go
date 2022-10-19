// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package artdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/oci/artdesc"
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
