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
	capriv, _, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	spec := &signutils.Specification{
		Subject: pkix.Name{
			CommonName: "ca-authority",
		},
		Validity:     10 * time.Hour,
		CAPrivateKey: capriv,
		IsCA:         true,
		Usages:       []interface{}{x509.ExtKeyUsageCodeSigning, x509.KeyUsageDigitalSignature},
	}

	ca, _, err := signutils.CreateCertificate(spec)
	Expect(err).To(Succeed())

	// leaf
	_, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	Context("direct", func() {
		spec := &signutils.Specification{
			Subject: pkix.Name{
				CommonName: "mandelsoft",
			},
			RootCAs:      ca,
			CAChain:      ca,
			CAPrivateKey: capriv,
			PublicKey:    pub,
			Usages:       []interface{}{x509.ExtKeyUsageCodeSigning, x509.KeyUsageDigitalSignature},
		}
		cert, _, err := signutils.CreateCertificate(spec)
		Expect(err).To(Succeed())

		pool := x509.NewCertPool()
		pool.AddCert(ca)

		It("identifies self-signed", func() {
			Expect(signutils.IsSelfSigned(ca)).To(BeTrue())
		})

		It("verifies for issuer", func() {
			err = signing.VerifyCert(nil, pool, "mandelsoft", cert)
			Expect(err).To(Succeed())
		})
		It("verifies for anonymous", func() {
			err = signing.VerifyCert(nil, pool, "", cert)
			Expect(err).To(Succeed())
		})
		It("fails for wrong issuer", func() {
			err = signing.VerifyCert(nil, pool, "x", cert)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("chain", func() {
		defer GinkgoRecover()

		interpriv, _, err := rsa.Handler{}.CreateKeyPair()
		Expect(err).To(Succeed())

		spec := &signutils.Specification{
			IsCA: true,
			Subject: pkix.Name{
				CommonName: "acme.org",
			},
			RootCAs:      ca,
			CAChain:      ca,
			CAPrivateKey: capriv,
			PublicKey:    interpriv,
			Usages:       []interface{}{x509.ExtKeyUsageCodeSigning, x509.KeyUsageDigitalSignature},
		}

		intercert, _, err := signutils.CreateCertificate(spec)
		Expect(err).To(Succeed())

		spec = &signutils.Specification{
			Subject: pkix.Name{
				CommonName:    "mandelsoft",
				Country:       []string{"DE", "US"},
				Locality:      []string{"Walldorf d"},
				StreetAddress: []string{"x y"},
				PostalCode:    []string{"69169"},
				Province:      []string{"BW"},
			},
			RootCAs:      ca,
			CAChain:      []*x509.Certificate{intercert, ca},
			CAPrivateKey: interpriv,
			PublicKey:    pub,
			Usages:       []interface{}{x509.ExtKeyUsageCodeSigning, x509.KeyUsageDigitalSignature},
		}
		cert, pemBytes, err := signutils.CreateCertificate(spec)
		Expect(err).To(Succeed())

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
			_, err = cert.Verify(opts)
			Expect(err).To(Succeed())
		})
	})
})
