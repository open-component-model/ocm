// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signutils_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/open-component-model/ocm/pkg/testutils"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/signutils"
)

var _ = Describe("normalization", func() {

	// root
	ca, capriv := Must2(rsa.CreateRootCertificate(signutils.CommonName("ca-authority"), 10*time.Hour))

	Context("direct", func() {
		cert, _, _ := Must3(rsa.CreateSigningCertificate(signutils.CommonName("mandelsoft"), ca, ca, capriv, 1*time.Hour))

		pool := x509.NewCertPool()
		pool.AddCert(ca)

		It("identifies self-signed", func() {
			Expect(signutils.IsSelfSigned(ca)).To(BeTrue())
		})

		It("verifies for issuer", func() {
			MustBeSuccessful(signing.VerifyCert(nil, pool, "mandelsoft", cert))
		})
		It("verifies for anonymous", func() {
			MustBeSuccessful(signing.VerifyCert(nil, pool, "", cert))
		})
		It("fails for wrong issuer", func() {
			MustFailWithMessage(signing.VerifyCert(nil, pool, "x", cert), `common name "mandelsoft" is invalid`)
		})
	})

	Context("chain", func() {
		defer GinkgoRecover()

		intercert, interBytes, interpriv := Must3(rsa.CreateSigningCertificate(signutils.CommonName("acme.org"), ca, ca, capriv, 1*time.Hour, true))

		cert, pemBytes, _ := Must3(rsa.CreateSigningCertificate(&pkix.Name{
			CommonName:    "mandelsoft",
			Country:       []string{"DE", "US"},
			Locality:      []string{"Walldorf d"},
			StreetAddress: []string{"x y"},
			PostalCode:    []string{"69169"},
			Province:      []string{"BW"},
		}, interBytes, ca, interpriv, 1*time.Hour))

		certs := Must(signutils.GetCertificateChain(pemBytes, false))
		Expect(len(certs)).To(Equal(3))

		pool := x509.NewCertPool()
		pool.AddCert(ca)

		interpool := x509.NewCertPool()
		interpool.AddCert(intercert)

		opts := x509.VerifyOptions{
			Intermediates: interpool,
			Roots:         pool,
			CurrentTime:   time.Now(),
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		}

		It("identifies non-self-signed", func() {
			Expect(signutils.IsSelfSigned(intercert)).To(BeFalse())
		})

		It("verifies", func() {
			fmt.Printf("%s\n", cert.Subject.String())
			MustBeSuccessful(cert.Verify(opts))
		})
	})
})
