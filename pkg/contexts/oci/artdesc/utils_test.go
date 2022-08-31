// Copyright 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
