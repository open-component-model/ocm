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

package rsakeypair_test

import (
	"bytes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/go-digest"

	"github.com/open-component-model/ocm/pkg/signing/handlers/rsa"

	. "github.com/open-component-model/ocm/cmds/ocm/testhelper"
)

const ISSUER = "mandelsoft"

var _ = Describe("Test Environment", func() {
	var env *TestEnv

	BeforeEach(func() {
		env = NewTestEnv()
	})

	AfterEach(func() {
		env.Cleanup()
	})

	It("create key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv")).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(`
created rsa key pair key.priv[key.pub]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.pub")
		Expect(err).To(Succeed())

		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(d.Hex(), 0, ISSUER, priv)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))
		Expect(sig.Issuer).To(Equal(ISSUER))

		err = rsa.Handler{}.Verify(d.Hex(), 0, sig, pub)
		Expect(err).To(Succeed())
	})

	It("create self-signed key pair", func() {

		buf := bytes.NewBuffer(nil)
		Expect(env.CatchOutput(buf).Execute("create", "rsakeypair", "key.priv", "CN=mandelsoft")).To(Succeed())
		Expect("\n" + buf.String()).To(Equal(`
created rsa key pair key.priv[key.cert]
`))
		priv, err := env.ReadFile("key.priv")
		Expect(err).To(Succeed())
		pub, err := env.ReadFile("key.cert")
		Expect(err).To(Succeed())

		d := digest.FromBytes([]byte("digest"))
		sig, err := rsa.Handler{}.Sign(d.Hex(), 0, "mandelsoft", priv)
		Expect(err).To(Succeed())
		Expect(sig.Algorithm).To(Equal(rsa.Algorithm))
		Expect(sig.MediaType).To(Equal(rsa.MediaType))
		Expect(sig.Issuer).To(Equal(ISSUER))

		err = rsa.Handler{}.Verify(d.Hex(), 0, sig, pub)
		Expect(err).To(Succeed())
	})
})
