// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/signing"
)

type DigestHandler struct {
	typ internal.DigesterType
}

var _ internal.BlobDigester = (*DigestHandler)(nil)

func (d *DigestHandler) GetType() internal.DigesterType {
	return d.typ
}

func (d *DigestHandler) DetermineDigest(resType string, meth internal.AccessMethod, preferred signing.Hasher) (*internal.DigestDescriptor, error) {
	return nil, nil
}

var _ = Describe("blob digester registry test", func() {
	var reg internal.BlobDigesterRegistry

	BeforeEach(func() {
		reg = internal.NewBlobDigesterRegistry()
	})

	It("copies registries", func() {
		mine := &DigestHandler{internal.DigesterType{
			HashAlgorithm:          "hash",
			NormalizationAlgorithm: "norm",
		}}

		reg.Register(mine, "arttype")

		h := reg.GetDigesterForType("arttype")
		Expect(h).To(Equal([]internal.BlobDigester{mine}))

		copy := reg.Copy()
		new := &DigestHandler{internal.DigesterType{
			HashAlgorithm:          "other",
			NormalizationAlgorithm: "norm",
		}}
		copy.Register(new, "arttype")

		h = reg.GetDigesterForType("arttype")
		Expect(h).To(Equal([]internal.BlobDigester{mine}))

		h = copy.GetDigesterForType("arttype")
		Expect(h).To(Equal([]internal.BlobDigester{mine, new}))

	})

})
