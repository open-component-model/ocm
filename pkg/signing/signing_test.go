// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package signing_test

import (
	"crypto/x509/pkix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/contexts/credentials"
	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"
)

var registry = signing.DefaultRegistry()

const NAME = "testsignature"

var ISSUER = &pkix.Name{CommonName: "mandelsoft"}

var _ = Describe("signing", func() {
	var defaultContext credentials.Context

	BeforeEach(func() {
		defaultContext = credentials.New()
	})

	It("uses rsa signer", func() {
		hasher := registry.GetHasher(sha256.Algorithm)
		hash, _ := signing.Hash(hasher.Create(), []byte("test"))

		priv, pub, err := rsa.Handler{}.CreateKeyPair()
		Expect(err).To(Succeed())

		registry.RegisterPublicKey(NAME, pub)
		registry.RegisterPrivateKey(NAME, priv)

		sctx := &signing.DefaultSigningContext{
			Hash:       hasher.Crypto(),
			PrivateKey: registry.GetPrivateKey(NAME),
			PublicKey:  pub,
			RootCerts:  nil,
			Issuer:     ISSUER,
		}
		sig, err := registry.GetSigner(rsa.Algorithm).Sign(defaultContext, hash, sctx)

		Expect(err).To(Succeed())
		Expect(sig.MediaType).To(Equal(rsa.MediaType))

		sctx.PublicKey = registry.GetPublicKey(NAME)
		Expect(registry.GetVerifier(rsa.Algorithm).Verify(hash, sig, sctx)).To(Succeed())
		hash = "A" + hash[1:]
		Expect(registry.GetVerifier(rsa.Algorithm).Verify(hash, sig, sctx)).To(HaveOccurred())
	})
})
