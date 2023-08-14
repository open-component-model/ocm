// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package compdesc_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/localblob"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/equivalent/testhelper"

	v1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
)

var _ = Describe("equivalence", func() {
	var labels v1.Labels
	var modtime *v1.Timestamp

	_ = modtime

	BeforeEach(func() {
		labels.Clear()
		labels.Set("label1", "value1", v1.WithSigning())
		labels.Set("label3", "value3")
	})

	Context("element meta", func() {
		var a, b *compdesc.ElementMeta

		BeforeEach(func() {
			a = &compdesc.ElementMeta{
				Name:          "r1",
				Version:       "v1",
				ExtraIdentity: v1.NewExtraIdentity("extra", "extra"),
				Labels:        labels.Copy(),
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
		})

		It("handles name change", func() {
			b.Name = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles version change", func() {
			b.Version = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles extra id change", func() {
			b.ExtraIdentity["X"] = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles non-volatile label change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
		})

		It("handles volatile label change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
		})
	})

	Context("resource", func() {
		var a, b *compdesc.Resource

		BeforeEach(func() {
			a = &compdesc.Resource{
				ResourceMeta: compdesc.ResourceMeta{
					ElementMeta: compdesc.ElementMeta{
						Name:   "r1",
						Labels: labels.Copy(),
					},
					Type:     "test",
					Relation: v1.LocalRelation,
					Digest: &v1.DigestSpec{
						HashAlgorithm:          "hash",
						NormalisationAlgorithm: "norm",
						Value:                  "x",
					},
				},
				Access: localblob.New("test", "test", "test", nil),
			}
			b = a.Copy()
		})

		It("handles equal", func() {
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles volatile meta change", func() {
			b.Labels[1].Value = []byte("X")
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})

		It("handles non-volatile meta change", func() {
			b.Labels[0].Value = []byte("X")
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles type change", func() {
			b.Type = "X"
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles relation change", func() {
			b.Relation = compdesc.ExternalRelation
			CheckNotLocalHashEqual(a.Equivalent(b))
			CheckNotLocalHashEqual(b.Equivalent(a))
		})

		It("handles access change", func() {
			b.Access = ociartifact.New("test")
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles undetectable digest", func() {
			b.Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
			a.Digest = nil
			CheckNotDetectable(a.Equivalent(b))
			CheckNotDetectable(b.Equivalent(a))
		})

		It("handles different digest", func() {
			b.Digest.Value = "X"
			CheckNotArtifactEqual(a.Equivalent(b))
			CheckNotArtifactEqual(b.Equivalent(a))
		})

	})

	Context("resources", func() {

	})

	Context("sources", func() {

	})

	Context("references", func() {

	})

	Context("signatures", func() {
		var s1, s2 *v1.Signature

		BeforeEach(func() {
			s1 = &v1.Signature{
				Name: "sig",
				Digest: v1.DigestSpec{
					HashAlgorithm:          "hash",
					NormalisationAlgorithm: "norm",
					Value:                  "H",
				},
				Signature: v1.SignatureSpec{
					Algorithm: "sign",
					Value:     "S",
					MediaType: "M",
					Issuer:    "issuer",
				},
			}
			s2 = s1.Copy()
			s2.Name = "other"
		})

		It("handles equal", func() {
			a := compdesc.Signatures{*s1}
			b := compdesc.Signatures{*s1.Copy()}
			CheckEquivalent(a.Equivalent(b))
			CheckEquivalent(b.Equivalent(a))
		})

		It("handles diff", func() {
			a := compdesc.Signatures{*s1}
			b := compdesc.Signatures{}
			CheckNotEquivalent(a.Equivalent(b))
			CheckNotEquivalent(b.Equivalent(a))
		})
	})

})
