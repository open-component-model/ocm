// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/v2/pkg/signing"
	"github.com/open-component-model/ocm/v2/pkg/signing/handlers/rsa"
)

var _ = Describe("normalization", func() {

	capriv, capub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	subject := pkix.Name{
		CommonName: "ca-authority",
	}
	caData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, capub, nil, capriv, true)
	Expect(err).To(Succeed())
	ca, err := x509.ParseCertificate(caData)
	Expect(err).To(Succeed())

	priv, pub, err := rsa.Handler{}.CreateKeyPair()
	Expect(err).To(Succeed())

	subject = pkix.Name{
		CommonName:    "mandelsoft",
		StreetAddress: []string{"some street 21"},
	}

	Context("foreignly signed", func() {
		certData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, pub, ca, capriv, false)
		Expect(err).To(Succeed())

		cert, err := x509.ParseCertificate(certData)
		Expect(err).To(Succeed())

		pool := x509.NewCertPool()
		pool.AddCert(ca)

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
	Context("self signed", func() {
		certData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, pub, nil, priv, false)
		Expect(err).To(Succeed())

		cert, err := x509.ParseCertificate(certData)
		Expect(err).To(Succeed())

		pool := x509.NewCertPool()
		pool.AddCert(cert)

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
})
