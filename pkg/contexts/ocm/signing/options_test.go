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
	. "github.com/open-component-model/ocm/pkg/contexts/ocm/signing"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

const NAME = "test"

var _ = Describe("options", func() {
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
	certData, err := signing.CreateCertificate(subject, nil, 10*time.Hour, pub, ca, capriv, false)
	Expect(err).To(Succeed())

	cert, err := x509.ParseCertificate(certData)
	Expect(err).To(Succeed())

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	It("verifies options for verification", func() {
		opts := NewOptions(
			RootCertificates(pool),
			VerifySignature(NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(signing.DefaultRegistry())).To(Succeed())
	})

	It("fails for options for verification without root cert", func() {
		opts := NewOptions(
			VerifySignature(NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(signing.DefaultRegistry())).To(HaveOccurred())
	})

	It("succeeds for options for signing with verification with root cert", func() {
		opts := NewOptions(
			RootCertificates(pool),
			Sign(signing.DefaultRegistry().GetSigner(rsa.Algorithm), NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(signing.DefaultRegistry())).To(Succeed())
	})

	It("fails for options for signing with verification without root cert", func() {
		opts := NewOptions(
			Sign(signing.DefaultRegistry().GetSigner(rsa.Algorithm), NAME),
			PrivateKey(NAME, priv),
			PublicKey(NAME, cert),
		)
		Expect(opts.Complete(signing.DefaultRegistry())).To(HaveOccurred())
	})
})
