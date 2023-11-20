// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signutils_test

import (
	"crypto/x509/pkix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

var _ = Describe("normalization", func() {
	Context("parse", func() {
		It("plain", func() {
			dn := Must(signutils.ParseDN("mandelsoft"))
			Expect(dn.String()).To(Equal("CN=mandelsoft"))
		})
		It("single field", func() {
			dn := Must(signutils.ParseDN("CN=mandelsoft"))
			Expect(dn.String()).To(Equal("CN=mandelsoft"))
		})
		It("two fields", func() {
			dn := Must(signutils.ParseDN("CN=mandelsoft,C=DE"))
			Expect(dn.String()).To(Equal("CN=mandelsoft,C=DE"))
		})
		It("three fields", func() {
			dn := Must(signutils.ParseDN("CN=mandelsoft,C=DE,ST=BW"))
			Expect(dn.String()).To(Equal("CN=mandelsoft,ST=BW,C=DE"))
		})
		It("double fields", func() {
			dn := Must(signutils.ParseDN("CN=mandelsoft,C=DE+C=US"))
			Expect(dn.String()).To(Equal("CN=mandelsoft,C=DE+C=US"))
		})
	})

	Context("match", func() {
		It("complete", func() {
			dn := pkix.Name{
				CommonName: "a",
				Country:    []string{"DE", "US"},
			}

			Expect(signutils.MatchDN(dn, dn)).NotTo(HaveOccurred())
		})
		It("partly", func() {
			dn := pkix.Name{
				CommonName: "a",
				Country:    []string{"DE", "US"},
			}

			p := dn
			p.Country = nil
			Expect(signutils.MatchDN(dn, p)).NotTo(HaveOccurred())
		})
		It("partly list", func() {
			dn := pkix.Name{
				CommonName: "a",
				Country:    []string{"DE", "US"},
			}

			p := dn
			p.Country = []string{"DE"}
			Expect(signutils.MatchDN(dn, p)).NotTo(HaveOccurred())
		})

		It("fails for missing", func() {
			dn := pkix.Name{
				CommonName: "a",
				Country:    []string{"DE", "US"},
			}

			p := dn
			p.Country = []string{"EG"}
			Expect(signutils.MatchDN(dn, p)).To(MatchError(`country "EG" not found`))
		})
	})
})
