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

package signing_test

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
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
