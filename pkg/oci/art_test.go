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

package oci_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-component-model/ocm/pkg/oci"
	"github.com/opencontainers/go-digest"
)

func CheckArt(ref string, exp *oci.ArtSpec) {
	spec, err := oci.ParseArt(ref)
	if exp == nil {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).To(Succeed())
		Expect(spec).To(Equal(*exp))
	}
}

var _ = Describe("art parsing", func() {
	digest := digest.Digest("sha256:3d05e105e350edf5be64fe356f4906dd3f9bf442a279e4142db9879bba8e677a")
	tag := "v1"

	It("succeeds", func() {
		CheckArt("ubuntu", &oci.ArtSpec{Repository: "ubuntu"})
		CheckArt("ubuntu/test", &oci.ArtSpec{Repository: "ubuntu/test"})
		CheckArt("ubuntu/test@"+digest.String(), &oci.ArtSpec{Repository: "ubuntu/test", Digest: &digest})
		CheckArt("ubuntu/test:"+tag, &oci.ArtSpec{Repository: "ubuntu/test", Tag: &tag})
		CheckArt("ubuntu/test:"+tag+"@"+digest.String(), &oci.ArtSpec{Repository: "ubuntu/test", Digest: &digest, Tag: &tag})
	})

	It("fails", func() {
		CheckArt("ubu@ntu", nil)
		CheckArt("ubu@sha256:123", nil)
	})

})
