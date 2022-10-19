// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
