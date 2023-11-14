// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package internal_test

import (
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/contexts/ocm"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/cpi/accspeccpi"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/internal"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

type DigestHandler struct {
	typ internal.DigesterType
}

var _ internal.BlobDigester = (*DigestHandler)(nil)

var digest = &internal.DigestDescriptor{
	HashAlgorithm:          "hash",
	NormalisationAlgorithm: "norm",
	Value:                  "Z",
}

func (d *DigestHandler) GetType() internal.DigesterType {
	return d.typ
}

func (d *DigestHandler) DetermineDigest(resType string, meth internal.AccessMethod, preferred signing.Hasher) (*internal.DigestDescriptor, error) {
	return digest, nil
}

var _ = Describe("blob digester registry test", func() {
	var reg internal.BlobDigesterRegistry
	var ext internal.BlobDigesterRegistry

	hasher := signing.DefaultRegistry().GetHasher(sha256.Algorithm)

	BeforeEach(func() {
		reg = internal.NewBlobDigesterRegistry()
		ext = internal.NewBlobDigesterRegistry(reg)
	})

	DescribeTable("copies registries",
		func(eff *internal.BlobDigesterRegistry) {
			mine := &DigestHandler{internal.DigesterType{
				HashAlgorithm:          "hash",
				NormalizationAlgorithm: "norm",
			}}
			reg.Register(mine, "arttype")

			h := (*eff).GetDigesterForType("arttype")
			Expect(h).To(Equal([]internal.BlobDigester{mine}))

			copy := (*eff).Copy()
			new := &DigestHandler{internal.DigesterType{
				HashAlgorithm:          "other",
				NormalizationAlgorithm: "norm",
			}}
			copy.Register(new, "arttype")

			h = (*eff).GetDigesterForType("arttype")
			Expect(h).To(Equal([]internal.BlobDigester{mine}))

			h = copy.GetDigesterForType("arttype")
			if *eff == ext {
				Expect(h).To(Equal([]internal.BlobDigester{new, mine}))
			} else {
				Expect(h).To(Equal([]internal.BlobDigester{mine, new}))
			}
		},
		Entry("plain", &reg),
		Entry("extend", &ext),
	)

	DescribeTable("uses digester to digest",
		func(eff *internal.BlobDigesterRegistry) {
			mine := &DigestHandler{internal.DigesterType{
				HashAlgorithm:          "hash",
				NormalizationAlgorithm: "norm",
			}}
			reg.Register(mine, "arttype")

			descs := Must((*eff).DetermineDigests("arttype", hasher, signing.DefaultRegistry(), NewDummyMethod()))
			Expect(descs).To(Equal([]internal.DigestDescriptor{*digest}))
		},
		Entry("plain", &reg),
		Entry("extend", &ext),
	)
})

type accessMethod struct {
}

var _ internal.AccessMethodImpl = (*accessMethod)(nil)

func (_ accessMethod) IsLocal() bool {
	return false
}

func (a accessMethod) GetKind() string {
	return "demo"
}

func (a accessMethod) AccessSpec() internal.AccessSpec {
	return nil
}

func (a accessMethod) Get() ([]byte, error) {
	return nil, nil
}

func (a accessMethod) Reader() (io.ReadCloser, error) {
	return nil, nil
}

func (a accessMethod) Close() error {
	return nil
}

func (a accessMethod) MimeType() string {
	return "application/demo"
}

func NewDummyMethod() ocm.AccessMethod {
	m, _ := accspeccpi.AccessMethodForImplementation(&accessMethod{}, nil)
	return m
}
