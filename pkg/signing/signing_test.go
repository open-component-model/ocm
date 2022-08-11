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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/open-component-model/ocm/pkg/signing/hasher/sha256"

	"github.com/open-component-model/ocm/pkg/signing"
	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"
)

var registry = signing.DefaultRegistry()

const NAME = "testsignature"

var _ = Describe("normalization", func() {

	It("Normalizes struct without excludes", func() {

		hasher := registry.GetHasher(sha256.Algorithm)
		hash, _ := signing.Hash(hasher.Create(), []byte("test"))

		priv, pub, err := rsa.Handler{}.CreateKeyPair()
		Expect(err).To(Succeed())

		registry.RegisterPublicKey(NAME, pub)
		registry.RegisterPrivateKey(NAME, priv)

		sig, err := registry.GetSigner(rsa.Algorithm).Sign(hash, hasher.Crypto(), "mandelsoft", registry.GetPrivateKey(NAME))

		Expect(err).To(Succeed())
		Expect(sig.MediaType).To(Equal(rsa.MediaType))
		Expect(sig.Issuer).To(Equal("mandelsoft"))

		Expect(registry.GetVerifier(rsa.Algorithm).Verify(hash, hasher.Crypto(), sig, registry.GetPublicKey(NAME))).To(Succeed())
		hash = "A" + hash[1:]
		Expect(registry.GetVerifier(rsa.Algorithm).Verify(hash, hasher.Crypto(), sig, registry.GetPublicKey(NAME))).To(HaveOccurred())
	})
})
