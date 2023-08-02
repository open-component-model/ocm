// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package transfer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/transfer"
	. "github.com/open-component-model/ocm/pkg/testutils"
)

var _ = Describe("basic merge operations for transport", func() {
	Context("merging labels", func() {
		var src metav1.Labels
		var dst metav1.Labels

		BeforeEach(func() {
			src.Set("add-signed", "new signed label", metav1.WithSigning())
			src.Set("add-unsigned", "new unsigned label")
			src.Set("old-signed", "old signed label", metav1.WithSigning())
			src.Set("old-unsigned", "old signed label")

			dst.Set("old-signed", "signed label", metav1.WithSigning())
			dst.Set("old-unsigned", "signed label")
			dst.Set("new-signed", "signed label", metav1.WithSigning())
			dst.Set("new-unsigned", "unsigned label")
		})

		It("add signed additional ones", func() {
			res := dst.Copy()
			MustBeSuccessful(transfer.MergeLabels(src, &res))
			Expect(res).To(ConsistOf(append(dst, *src.GetDef("add-unsigned"))))
		})

	})

	////////////////////////////////////////////////////////////////////////////
	Context("merging signatures", func() {
		var src metav1.Signatures
		var dst metav1.Signatures

		BeforeEach(func() {
			src.Set(metav1.Signature{
				Name: "add",
				Digest: metav1.DigestSpec{
					Value: "add",
				},
				Signature: metav1.SignatureSpec{
					Value: "add-sig",
				},
			})
			src.Set(metav1.Signature{
				Name: "old",
				Digest: metav1.DigestSpec{
					Value: "old",
				},
				Signature: metav1.SignatureSpec{
					Value: "old-sig",
				},
			})

			dst.Set(metav1.Signature{
				Name: "new",
				Digest: metav1.DigestSpec{
					Value: "new",
				},
				Signature: metav1.SignatureSpec{
					Value: "new-sig",
				},
			})
			dst.Set(metav1.Signature{
				Name: "old",
				Digest: metav1.DigestSpec{
					Value: "old",
				},
				Signature: metav1.SignatureSpec{
					Value: "mod-sig",
				},
			})
		})

		It("add signed additional ones", func() {
			res := dst.Copy()
			MustBeSuccessful(transfer.MergeSignatures(src, &res))
			Expect(res).To(ConsistOf(append(dst, *src.GetByName("add"))))
		})
	})

	////////////////////////////////////////////////////////////////////////////
	Context("merging component descriptors", func() {

	})
})
