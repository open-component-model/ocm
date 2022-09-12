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

package core_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/core"
	"github.com/open-component-model/ocm/pkg/signing"
)

type DigestHandler struct {
	typ core.DigesterType
}

var _ core.BlobDigester = (*DigestHandler)(nil)

func (d *DigestHandler) GetType() core.DigesterType {
	return d.typ
}

func (d *DigestHandler) DetermineDigest(resType string, meth core.AccessMethod, preferred signing.Hasher) (*core.DigestDescriptor, error) {
	return nil, nil
}

var _ = Describe("blob digester registry test", func() {
	var reg core.BlobDigesterRegistry

	BeforeEach(func() {
		reg = core.NewBlobDigesterRegistry()
	})

	It("copies registries", func() {
		mine := &DigestHandler{core.DigesterType{
			HashAlgorithm:          "hash",
			NormalizationAlgorithm: "norm",
		}}

		reg.Register(mine, "arttype")

		h := reg.GetDigesterForType("arttype")
		Expect(h).To(Equal([]core.BlobDigester{mine}))

		copy := reg.Copy()
		new := &DigestHandler{core.DigesterType{
			HashAlgorithm:          "other",
			NormalizationAlgorithm: "norm",
		}}
		copy.Register(new, "arttype")

		h = reg.GetDigesterForType("arttype")
		Expect(h).To(Equal([]core.BlobDigester{mine}))

		h = copy.GetDigesterForType("arttype")
		Expect(h).To(Equal([]core.BlobDigester{mine, new}))

	})

})
