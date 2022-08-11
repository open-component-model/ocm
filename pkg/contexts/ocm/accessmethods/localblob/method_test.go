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

package localblob_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociblob"
	"github.com/open-component-model/ocm/pkg/mime"
	"github.com/open-component-model/ocm/pkg/runtime"
)

const OCIPATH = "/tmp/oci"
const OCINAMESPACE = "ocm/test"
const OCIVERSION = "v2.0"
const OCIHOST = "alias"

var _ = Describe("Method", func() {

	var data = `globalAccess:
  digest: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
  mediaType: application/tar+gzip
  ref: ghcr.io/vasu1124/ocm/component-descriptors/github.com/vasu1124/introspect-delivery
  size: 11287
  type: ociBlob
localReference: sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a
mediaType: application/tar+gzip
type: localBlob
`
	_ = data

	It("marshal/unmarshal simple", func() {
		spec := localblob.New("path", "hint", mime.MIME_TEXT, nil)
		data, err := json.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(string(data)).To(Equal("{\"type\":\"localBlob\",\"localReference\":\"path\",\"mediaType\":\"text/plain\",\"referenceName\":\"hint\"}"))
		var r localblob.AccessSpec
		Expect(json.Unmarshal(data, &r)).To(Succeed())
		Expect(&r).To(Equal(spec))
	})

	It("marshal/unmarshal with global", func() {
		spec := &localblob.AccessSpec{}
		Expect(runtime.DefaultYAMLEncoding.Unmarshal([]byte(data), spec)).To(Succeed())

		r, err := runtime.DefaultYAMLEncoding.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(string(r)).To(Equal(data))

		global := ociblob.New(
			"ghcr.io/vasu1124/ocm/component-descriptors/github.com/vasu1124/introspect-delivery",
			"sha256:1bf729fa00e355199e711933ccfa27467ee3d2de1343aef2a7c1ecbdf885e63a",
			"application/tar+gzip",
			11287,
		)
		Expect(spec.GlobalAccess.Evaluate(ocm.DefaultContext())).To(Equal(global))

		r, err = runtime.DefaultYAMLEncoding.Marshal(spec)
		Expect(err).To(Succeed())
		Expect(string(r)).To(Equal(data))
	})

})
